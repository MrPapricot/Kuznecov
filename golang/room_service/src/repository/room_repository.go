package repository

import (
	db_adapter "db_adapter/src"
	"errors"
	"fmt"
	"time"

	user_models "shared/user_data"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	NoUserFound   = errors.New("no user found for given token")
	TokenExpired  = errors.New("token expired")
	InvalidToken  = errors.New("invalid token")
	DatabaseError = errors.New("database error")
	RoomNotFound  = errors.New("room not found")
	NotRoomMember = errors.New("user is not a member of the room")
)

type RoomRepository struct {
	db            *db_adapter.PostgresAdapter
	jwtSecret     string
	jwtExpireTime time.Duration
}

func NewRoomRepository(db *db_adapter.PostgresAdapter, jwtSecret string, jwtExpireTime time.Duration) RoomRepository {
	return RoomRepository{db: db, jwtSecret: jwtSecret, jwtExpireTime: jwtExpireTime}
}

// getAndValidateUUIDFromToken парсит токен И проверяет существование пользователя в БД
func (r *RoomRepository) getAndValidateUUIDFromToken(tokenString string) (uuid.UUID, error) {
	// 1. Парсим и проверяем подпись/срок действия JWT
	userUUID, err := r.getUUIDFromToken(tokenString)
	if err != nil {
		return uuid.Nil, err // Вернет InvalidToken или TokenExpired
	}

	// 2. Проверяем, что пользователь с таким UUID действительно существует в БД
	exists, err := r.db.UserExists(userUUID)
	if err != nil {
		return uuid.Nil, DatabaseError
	}

	if !exists {
		return uuid.Nil, NoUserFound
	}

	return userUUID, nil
}

func (r *RoomRepository) getUUIDFromToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &user_models.TokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(r.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, TokenExpired
		}
		return uuid.Nil, InvalidToken
	}

	if !token.Valid {
		return uuid.Nil, InvalidToken
	}

	claims, ok := token.Claims.(*user_models.TokenClaims)
	if !ok || !token.Valid {
		return uuid.Nil, InvalidToken
	}

	return claims.UUID, nil
}

func (r *RoomRepository) CreateRoom(token, name, description string) (uuid.UUID, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return uuid.Nil, err
	}

	roomUUID, err := r.db.CreateRoom(db_adapter.CreateRoomBody{
		Name: name, Description: description, OwnerUUID: userUUID,
	})
	if err != nil {
		return uuid.Nil, DatabaseError
	}
	return roomUUID, nil
}

func (r *RoomRepository) AddMembers(token string, roomUUID uuid.UUID, usernames []string) (members []db_adapter.MemberInfo, notFound []string, alreadyMembers []string, err error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return nil, nil, nil, err
	}

	room, err := r.db.GetRoomByUUID(roomUUID)
	if err != nil || room.OwnerUUID != userUUID {
		// Скрываем существование комнаты от не-владельцев, как требуется
		return nil, nil, nil, RoomNotFound
	}

	usernameToUUID, err := r.db.GetUserUUIDsByUsername(usernames)
	if err != nil {
		return nil, nil, nil, DatabaseError
	}

	notFound = []string{}
	memberUUIDs := []uuid.UUID{}
	for _, username := range usernames {
		if uuid, ok := usernameToUUID[username]; ok {
			memberUUIDs = append(memberUUIDs, uuid)
		} else {
			notFound = append(notFound, username)
		}
	}

	alreadyMemberUUIDs, err := r.db.AddRoomMembers(roomUUID, memberUUIDs)
	if err != nil {
		return nil, nil, nil, DatabaseError
	}

	uuidToUsername := make(map[uuid.UUID]string)
	for username, uuid := range usernameToUUID {
		uuidToUsername[uuid] = username
	}
	for _, uuid := range alreadyMemberUUIDs {
		if username, ok := uuidToUsername[uuid]; ok {
			alreadyMembers = append(alreadyMembers, username)
		}
	}

	members, err = r.db.GetRoomMembers(roomUUID)
	if err != nil {
		return nil, nil, nil, DatabaseError
	}

	return members, notFound, alreadyMembers, nil
}

func (r *RoomRepository) GetRoomInfo(token string, roomUUID uuid.UUID) (db_adapter.RoomInfo, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return db_adapter.RoomInfo{}, err
	}

	isMember, err := r.db.CheckUserIsRoomMember(roomUUID, userUUID)
	if err != nil {
		return db_adapter.RoomInfo{}, DatabaseError
	}
	if !isMember {
		return db_adapter.RoomInfo{}, NotRoomMember
	}

	info, err := r.db.GetRoomInfo(roomUUID)
	if err != nil {
		if errors.Is(err, db_adapter.ErrRoomNotFound) {
			return db_adapter.RoomInfo{}, RoomNotFound
		}
		return db_adapter.RoomInfo{}, DatabaseError
	}
	return info, nil
}

func (r *RoomRepository) RemoveMembers(token string, roomUUID uuid.UUID, usernames []string) (members []db_adapter.MemberInfo, notFound []string, ownerError string, err error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return nil, nil, "", err
	}

	room, err := r.db.GetRoomByUUID(roomUUID)
	if err != nil || room.OwnerUUID != userUUID {
		return nil, nil, "", RoomNotFound
	}

	usernameToUUID, err := r.db.GetUserUUIDsByUsername(usernames)
	if err != nil {
		return nil, nil, "", DatabaseError
	}

	notFound = []string{}
	memberUUIDs := []uuid.UUID{}

	for _, username := range usernames {
		if targetUUID, ok := usernameToUUID[username]; ok {
			if targetUUID == room.OwnerUUID {
				ownerError = fmt.Sprintf("user %s is owner and can not be excluded from room members", username)
				continue // Продолжаем удалять остальных
			}
			memberUUIDs = append(memberUUIDs, targetUUID)
		} else {
			notFound = append(notFound, username)
		}
	}

	_, _, err = r.db.RemoveRoomMembers(roomUUID, memberUUIDs)
	if err != nil {
		return nil, nil, "", DatabaseError
	}

	members, err = r.db.GetRoomMembers(roomUUID)
	if err != nil {
		return nil, nil, "", DatabaseError
	}

	return members, notFound, ownerError, nil
}

func (r *RoomRepository) UpdateRoomName(token string, roomUUID uuid.UUID, newName string) (string, string, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return "", "", err
	}

	isOwner, err := r.db.CheckUserIsRoomOwner(roomUUID, userUUID)
	if err != nil || !isOwner {
		return "", "", RoomNotFound
	}

	oldName, err := r.db.UpdateRoomName(roomUUID, newName)
	if err != nil {
		if errors.Is(err, db_adapter.ErrRoomNotFound) {
			return "", "", RoomNotFound
		}
		return "", "", DatabaseError
	}
	return oldName, newName, nil
}

func (r *RoomRepository) UpdateRoomDescription(token string, roomUUID uuid.UUID, newDesc string) (string, string, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return "", "", err
	}

	isOwner, err := r.db.CheckUserIsRoomOwner(roomUUID, userUUID)
	if err != nil || !isOwner {
		return "", "", RoomNotFound
	}

	oldDesc, err := r.db.UpdateRoomDescription(roomUUID, newDesc)
	if err != nil {
		if errors.Is(err, db_adapter.ErrRoomNotFound) {
			return "", "", RoomNotFound
		}
		return "", "", DatabaseError
	}
	return oldDesc, newDesc, nil
}

func (r *RoomRepository) DeleteRoom(token string, roomUUID uuid.UUID) error {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return err
	}

	isOwner, err := r.db.CheckUserIsRoomOwner(roomUUID, userUUID)
	if err != nil || !isOwner {
		return RoomNotFound
	}

	err = r.db.DeleteRoom(roomUUID)
	if err != nil {
		if errors.Is(err, db_adapter.ErrRoomNotFound) {
			return RoomNotFound
		}
		return DatabaseError
	}
	return nil
}

func (r *RoomRepository) GetOwnedRooms(token string) ([]db_adapter.Room, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return nil, err
	}
	rooms, err := r.db.GetRoomsByOwner(userUUID)
	if err != nil {
		return nil, DatabaseError
	}
	return rooms, nil
}

func (r *RoomRepository) GetJoinedRooms(token string) ([]db_adapter.Room, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return nil, err
	}
	rooms, err := r.db.GetRoomsByMember(userUUID)
	if err != nil {
		return nil, DatabaseError
	}
	return rooms, nil
}

func (r *RoomRepository) HealthCheck() error {
	return r.db.HealthCheck()
}
