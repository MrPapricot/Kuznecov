package db_adapter

import (
	"errors"
)

// Кастомные ошибки адаптера
var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrDatabaseError         = errors.New("database error")
	ErrTransactionError      = errors.New("failed to start transaction")
	ErrRoomNotFound          = errors.New("room not found")
	ErrCharacterNotFound     = errors.New("character not found")
	ErrInvalidStatSum        = errors.New("invalid stat points sum")
	ErrInvalidStatValue      = errors.New("stat value must be between 0 and 10")
	ErrInvalidLevel          = errors.New("level must be between 1 and 21")
	ErrInvalidStatName       = errors.New("invalid stat name")
	ErrStatMaxedOrLevelMaxed = errors.New("stat is already 10 or character level is 21")
)

// IsUniqueViolationError проверяет, является ли ошибка ошибкой уникальности
func IsUniqueViolationError(err error) bool {
	return errors.Is(err, ErrEmailAlreadyExists) || errors.Is(err, ErrUsernameAlreadyExists)
}
