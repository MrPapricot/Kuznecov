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
	NoUserFound           = errors.New("no user found for given token")
	TokenExpired          = errors.New("token expired")
	InvalidToken          = errors.New("invalid token")
	DatabaseError         = errors.New("database error")
	CharacterNotFound     = errors.New("character not found")
	NotCharacterOwner     = errors.New("user is not the owner of this character")
	InvalidStatSum        = errors.New("invalid stat points sum")
	InvalidStatValue      = errors.New("stat value must be between 0 and 10")
	InvalidLevel          = errors.New("level must be between 1 and 21")
	InvalidStatName       = errors.New("invalid stat name")
	StatMaxedOrLevelMaxed = errors.New("stat is already 10 or character level is 21")
)

type CharacterRepository struct {
	db            *db_adapter.PostgresAdapter
	jwtSecret     string
	jwtExpireTime time.Duration
}

func NewCharacterRepository(db *db_adapter.PostgresAdapter, jwtSecret string, jwtExpireTime time.Duration) CharacterRepository {
	return CharacterRepository{db: db, jwtSecret: jwtSecret, jwtExpireTime: jwtExpireTime}
}

func (r *CharacterRepository) getAndValidateUUIDFromToken(tokenString string) (uuid.UUID, error) {
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

	// Проверяем существование пользователя в БД
	exists, err := r.db.UserExists(claims.UUID)
	if err != nil {
		return uuid.Nil, DatabaseError
	}
	if !exists {
		return uuid.Nil, NoUserFound
	}

	return claims.UUID, nil
}

func (r *CharacterRepository) CreateCharacter(token, name, description string, level int16, stats db_adapter.CreateCharacterBody) (uuid.UUID, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return uuid.Nil, err
	}

	if level < 1 || level > 21 {
		return uuid.Nil, InvalidLevel
	}

	// Проверка значений статов (0-10)
	statsArr := []int16{stats.Strength, stats.Perception, stats.Endurance, stats.Charisma, stats.Intelligence, stats.Agility, stats.Luck}
	sum := int16(0)
	for _, stat := range statsArr {
		if stat < 0 || stat > 10 {
			return uuid.Nil, InvalidStatValue
		}
		sum += stat
	}

	expectedSum := int16(40 + (level - 1))
	if sum != expectedSum {
		return uuid.Nil, InvalidStatSum
	}

	stats.Name = name
	stats.CharacterDescription = description
	stats.Level = level
	stats.OwnerUUID = userUUID

	charUUID, err := r.db.CreateCharacter(stats)
	if err != nil {
		return uuid.Nil, DatabaseError
	}

	return charUUID, nil
}

func (r *CharacterRepository) UpdateCharacterName(token string, charUUID uuid.UUID, newName string) (string, string, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return "", "", err
	}

	isOwner, err := r.db.CheckUserIsCharacterOwner(charUUID, userUUID)
	if err != nil || !isOwner {
		return "", "", NotCharacterOwner
	}

	oldName, err := r.db.UpdateCharacterName(charUUID, newName)
	if err != nil {
		if errors.Is(err, db_adapter.ErrCharacterNotFound) {
			return "", "", CharacterNotFound
		}
		return "", "", DatabaseError
	}
	return oldName, newName, nil
}

func (r *CharacterRepository) UpdateCharacterDescription(token string, charUUID uuid.UUID, newDesc string) (string, string, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return "", "", err
	}

	isOwner, err := r.db.CheckUserIsCharacterOwner(charUUID, userUUID)
	if err != nil || !isOwner {
		return "", "", NotCharacterOwner
	}

	oldDesc, err := r.db.UpdateCharacterDescription(charUUID, newDesc)
	if err != nil {
		if errors.Is(err, db_adapter.ErrCharacterNotFound) {
			return "", "", CharacterNotFound
		}
		return "", "", DatabaseError
	}
	return oldDesc, newDesc, nil
}

func (r *CharacterRepository) DeleteCharacter(token string, charUUID uuid.UUID) error {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return err
	}

	isOwner, err := r.db.CheckUserIsCharacterOwner(charUUID, userUUID)
	if err != nil || !isOwner {
		return NotCharacterOwner
	}

	err = r.db.DeleteCharacter(charUUID)
	if err != nil {
		if errors.Is(err, db_adapter.ErrCharacterNotFound) {
			return CharacterNotFound
		}
		return DatabaseError
	}
	return nil
}

func (r *CharacterRepository) LevelUpCharacter(token string, charUUID uuid.UUID, statName string) (int16, int16, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return 0, 0, err
	}

	isOwner, err := r.db.CheckUserIsCharacterOwner(charUUID, userUUID)
	if err != nil || !isOwner {
		return 0, 0, NotCharacterOwner
	}

	newLevel, newStat, err := r.db.LevelUpStat(charUUID, statName)
	if err != nil {
		if errors.Is(err, db_adapter.ErrInvalidStatName) {
			return 0, 0, InvalidStatName
		}
		if errors.Is(err, db_adapter.ErrStatMaxedOrLevelMaxed) {
			return 0, 0, StatMaxedOrLevelMaxed
		}
		return 0, 0, DatabaseError
	}

	return newLevel, newStat, nil
}

func (r *CharacterRepository) GetCharacterInfo(token string, charUUID uuid.UUID) (db_adapter.Character, error) {
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return db_adapter.Character{}, err
	}

	isOwner, err := r.db.CheckUserIsCharacterOwner(charUUID, userUUID)
	if err != nil || !isOwner {
		return db_adapter.Character{}, NotCharacterOwner
	}

	char, err := r.db.GetCharacterByUUID(charUUID)
	if err != nil {
		if errors.Is(err, db_adapter.ErrCharacterNotFound) {
			return db_adapter.Character{}, CharacterNotFound
		}
		return db_adapter.Character{}, DatabaseError
	}
	return char, nil
}

func (r *CharacterRepository) HealthCheck() error {
	return r.db.HealthCheck()
}
