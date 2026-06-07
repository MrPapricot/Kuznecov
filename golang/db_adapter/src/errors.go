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
)

// IsUniqueViolationError проверяет, является ли ошибка ошибкой уникальности
func IsUniqueViolationError(err error) bool {
	return errors.Is(err, ErrEmailAlreadyExists) || errors.Is(err, ErrUsernameAlreadyExists)
}
