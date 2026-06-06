package repository

import (
	"errors"
	"fmt"

	db_adapter "db_adapter/src"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

// Кастомные ошибки репозитория
var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
)

type UserRepository struct {
	adapter *db_adapter.PostgresAdapter
}

func NewUserRepository(adapter *db_adapter.PostgresAdapter) *UserRepository {
	return &UserRepository{adapter: adapter}
}

// Create создаёт пользователя и возвращает его UUID
func (r *UserRepository) Create(username, email, password string) (uuid.UUID, error) {
	// Хешируем пароль
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to hash password: %w", err)
	}

	body := db_adapter.CreateUserBody{
		UserName:     username,
		Email:        email,
		PasswordHash: hash,
	}

	userID, err := r.adapter.CreateUser(body)
	if err != nil {
		// Транслируем ошибки адаптера в ошибки репозитория
		switch {
		case errors.Is(err, db_adapter.ErrEmailAlreadyExists):
			return uuid.Nil, ErrEmailAlreadyExists
		case errors.Is(err, db_adapter.ErrUsernameAlreadyExists):
			return uuid.Nil, ErrUsernameAlreadyExists
		default:
			return uuid.Nil, err
		}
	}

	return userID, nil
}

// GetUserByEmail возвращает UUID и хеш пароля пользователя
func (r *UserRepository) GetUserByEmail(email string) (uuid.UUID, string, error) {
	userID, passwordHash, err := r.adapter.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, db_adapter.ErrUserNotFound) {
			return uuid.Nil, "", ErrUserNotFound
		}
		return uuid.Nil, "", err
	}
	return userID, passwordHash, nil
}

// ComparePasswords проверяет, соответствует ли пароль хешу
func (r *UserRepository) ComparePasswords(password, passwordHash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, passwordHash)
	if err != nil {
		return false, fmt.Errorf("failed to compare password: %w", err)
	}
	return match, nil
}

func (r *UserRepository) HealthCheck() error {
	return r.adapter.HealthCheck()
}
