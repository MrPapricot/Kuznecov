package db_adapter

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Room struct {
	UUID        uuid.UUID
	Name        string
	Description sql.NullString
	OwnerUUID   uuid.UUID
	CreatedAt   time.Time
}

type MemberInfo struct {
	UUID     uuid.UUID
	Username string
	JoinedAt time.Time
}

type RoomInfo struct {
	UUID        uuid.UUID
	Name        string
	Description string
	OwnerUUID   uuid.UUID
	OwnerName   string
	CreatedAt   time.Time
	Members     []MemberInfo
}

type CreateRoomBody struct {
	Name        string
	Description string
	OwnerUUID   uuid.UUID
}

func (adapter *PostgresAdapter) CreateRoom(body CreateRoomBody) (uuid.UUID, error) {
	tx, err := adapter.db.Begin()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrTransactionError, err)
	}
	defer tx.Rollback()

	var roomUUID uuid.UUID
	err = tx.QueryRow(`
		INSERT INTO rooms (name, room_description, owner_uuid)
		VALUES ($1, $2, $3)
		RETURNING uuid
	`, body.Name, nullString(body.Description), body.OwnerUUID).Scan(&roomUUID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	_, err = tx.Exec(`
		INSERT INTO rooms_users_relations (room_uuid, member_uuid)
		VALUES ($1, $2)
	`, roomUUID, body.OwnerUUID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrTransactionError, err)
	}

	return roomUUID, nil
}

func (adapter *PostgresAdapter) GetRoomByUUID(roomUUID uuid.UUID) (Room, error) {
	var room Room
	err := adapter.db.QueryRow(`
		SELECT uuid, name, room_description, owner_uuid, created_at
		FROM rooms WHERE uuid = $1
	`, roomUUID).Scan(&room.UUID, &room.Name, &room.Description, &room.OwnerUUID, &room.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Room{}, ErrRoomNotFound
		}
		return Room{}, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return room, nil
}

func (adapter *PostgresAdapter) GetRoomInfo(roomUUID uuid.UUID) (RoomInfo, error) {
	var info RoomInfo
	err := adapter.db.QueryRow(`
		SELECT r.uuid, r.name, COALESCE(r.room_description, ''), r.owner_uuid, u.username, r.created_at
		FROM rooms r
		JOIN users u ON r.owner_uuid = u.uuid
		WHERE r.uuid = $1
	`, roomUUID).Scan(&info.UUID, &info.Name, &info.Description, &info.OwnerUUID, &info.OwnerName, &info.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RoomInfo{}, ErrRoomNotFound
		}
		return RoomInfo{}, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	rows, err := adapter.db.Query(`
		SELECT u.uuid, u.username, rel.joined_at
		FROM rooms_users_relations rel
		JOIN users u ON rel.member_uuid = u.uuid
		WHERE rel.room_uuid = $1
		ORDER BY rel.joined_at
	`, roomUUID)
	if err != nil {
		return RoomInfo{}, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var member MemberInfo
		if err := rows.Scan(&member.UUID, &member.Username, &member.JoinedAt); err != nil {
			return RoomInfo{}, fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}
		info.Members = append(info.Members, member)
	}

	return info, nil
}

func (adapter *PostgresAdapter) CheckUserIsRoomOwner(roomUUID, userUUID uuid.UUID) (bool, error) {
	var isOwner bool
	err := adapter.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM rooms WHERE uuid = $1 AND owner_uuid = $2)
	`, roomUUID, userUUID).Scan(&isOwner)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return isOwner, nil
}

func (adapter *PostgresAdapter) CheckUserIsRoomMember(roomUUID, userUUID uuid.UUID) (bool, error) {
	var isMember bool
	err := adapter.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM rooms_users_relations WHERE room_uuid = $1 AND member_uuid = $2)
	`, roomUUID, userUUID).Scan(&isMember)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return isMember, nil
}

// GetUserUUIDsByUsername возвращает map[username]uuid для списка username
func (adapter *PostgresAdapter) GetUserUUIDsByUsername(usernames []string) (map[string]uuid.UUID, error) {
	if len(usernames) == 0 {
		return map[string]uuid.UUID{}, nil
	}

	// Используем ANY($1) и оборачиваем слайс в pq.Array
	query := `SELECT username, uuid FROM users WHERE username = ANY($1)`

	rows, err := adapter.db.Query(query, pq.Array(usernames))
	if err != nil {
		fmt.Printf("Database error: %#v\n", err)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	defer rows.Close()

	result := make(map[string]uuid.UUID)
	for rows.Next() {
		var username string
		var userUUID uuid.UUID
		if err := rows.Scan(&username, &userUUID); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}
		result[username] = userUUID
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return result, nil
}

func (adapter *PostgresAdapter) AddRoomMembers(roomUUID uuid.UUID, memberUUIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(memberUUIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	tx, err := adapter.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransactionError, err)
	}
	defer tx.Rollback()

	alreadyMembers := []uuid.UUID{}
	for _, memberUUID := range memberUUIDs {
		var exists bool
		err := tx.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM rooms_users_relations WHERE room_uuid = $1 AND member_uuid = $2)
		`, roomUUID, memberUUID).Scan(&exists)
		if err != nil {
			fmt.Printf("%#v: %#v", ErrDatabaseError, err)
			return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}

		if exists {
			alreadyMembers = append(alreadyMembers, memberUUID)
			continue
		}

		_, err = tx.Exec(`INSERT INTO rooms_users_relations (room_uuid, member_uuid) VALUES ($1, $2)`, roomUUID, memberUUID)
		if err != nil {
			fmt.Printf("%#v: %#v", ErrDatabaseError, err)
			return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransactionError, err)
	}
	return alreadyMembers, nil
}

func (adapter *PostgresAdapter) RemoveRoomMembers(roomUUID uuid.UUID, memberUUIDs []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error) {
	if len(memberUUIDs) == 0 {
		return []uuid.UUID{}, []uuid.UUID{}, nil
	}

	removed := []uuid.UUID{}
	notFound := []uuid.UUID{}

	for _, memberUUID := range memberUUIDs {
		result, err := adapter.db.Exec(`
			DELETE FROM rooms_users_relations WHERE room_uuid = $1 AND member_uuid = $2
		`, roomUUID, memberUUID)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			notFound = append(notFound, memberUUID)
		} else {
			removed = append(removed, memberUUID)
		}
	}
	return removed, notFound, nil
}

func (adapter *PostgresAdapter) GetRoomMembers(roomUUID uuid.UUID) ([]MemberInfo, error) {
	rows, err := adapter.db.Query(`
		SELECT u.uuid, u.username, rel.joined_at
		FROM rooms_users_relations rel
		JOIN users u ON rel.member_uuid = u.uuid
		WHERE rel.room_uuid = $1
		ORDER BY rel.joined_at
	`, roomUUID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	defer rows.Close()

	var members []MemberInfo
	for rows.Next() {
		var member MemberInfo
		if err := rows.Scan(&member.UUID, &member.Username, &member.JoinedAt); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}
		members = append(members, member)
	}
	return members, nil
}

func (adapter *PostgresAdapter) UpdateRoomName(roomUUID uuid.UUID, newName string) (string, error) {
	var oldName string

	// 1. Сначала получаем текущее (старое) имя
	err := adapter.db.QueryRow(`SELECT name FROM rooms WHERE uuid = $1`, roomUUID).Scan(&oldName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrRoomNotFound
		}
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// 2. Затем выполняем обновление
	_, err = adapter.db.Exec(`UPDATE rooms SET name = $1 WHERE uuid = $2`, newName, roomUUID)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return oldName, nil
}

// UpdateRoomDescription обновляет описание и возвращает СТАРОЕ описание
func (adapter *PostgresAdapter) UpdateRoomDescription(roomUUID uuid.UUID, newDesc string) (string, error) {
	var oldDesc sql.NullString

	// 1. Сначала получаем текущее (старое) описание
	err := adapter.db.QueryRow(`SELECT room_description FROM rooms WHERE uuid = $1`, roomUUID).Scan(&oldDesc)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrRoomNotFound
		}
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// 2. Затем выполняем обновление
	_, err = adapter.db.Exec(`UPDATE rooms SET room_description = $1 WHERE uuid = $2`, nullString(newDesc), roomUUID)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return oldDesc.String, nil
}

func (adapter *PostgresAdapter) DeleteRoom(roomUUID uuid.UUID) error {
	result, err := adapter.db.Exec(`DELETE FROM rooms WHERE uuid = $1`, roomUUID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrRoomNotFound
	}
	return nil
}

func (adapter *PostgresAdapter) GetRoomsByOwner(ownerUUID uuid.UUID) ([]Room, error) {
	rows, err := adapter.db.Query(`SELECT uuid, name, room_description, owner_uuid, created_at FROM rooms WHERE owner_uuid = $1 ORDER BY created_at DESC`, ownerUUID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.UUID, &room.Name, &room.Description, &room.OwnerUUID, &room.CreatedAt); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (adapter *PostgresAdapter) GetRoomsByMember(memberUUID uuid.UUID) ([]Room, error) {
	rows, err := adapter.db.Query(`
		SELECT r.uuid, r.name, r.room_description, r.owner_uuid, r.created_at
		FROM rooms r JOIN rooms_users_relations rel ON r.uuid = rel.room_uuid
		WHERE rel.member_uuid = $1 ORDER BY r.created_at DESC
	`, memberUUID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.UUID, &room.Name, &room.Description, &room.OwnerUUID, &room.CreatedAt); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
