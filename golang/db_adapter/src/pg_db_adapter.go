package db_adapter

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type CreateUserBody struct {
	UserName     string
	Email        string
	PasswordHash string
}

type PostgresAdapter struct {
	db *sqlx.DB
}

type PostgresConnectOptions struct {
	Host         string
	Port         uint16
	UserName     string
	UserPassword string
	DBName       string
}

func PostgresConnect(options PostgresConnectOptions) (PostgresAdapter, error) {
	connect_url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		options.UserName, options.UserPassword, options.Host, options.Port, options.DBName)

	db, err := sqlx.Connect("postgres", connect_url)
	if err != nil {
		return PostgresAdapter{nil}, err
	}

	return PostgresAdapter{db}, nil
}

// CreateUser создаёт пользователя в БД
func (adapter *PostgresAdapter) CreateUser(body CreateUserBody) (uuid.UUID, error) {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING uuid
	`

	var userID uuid.UUID
	err := adapter.db.QueryRow(query, body.UserName, body.Email, body.PasswordHash).Scan(&userID)
	if err != nil {
		// Обрабатываем специфические ошибки PostgreSQL
		if parsedErr := adapter.parseInsertError(err); parsedErr != nil {
			return uuid.Nil, parsedErr
		}
		return uuid.Nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return userID, nil
}

// GetUserByEmail возвращает UUID и хеш пароля пользователя по email
func (adapter *PostgresAdapter) GetUserByEmail(email string) (uuid.UUID, string, error) {
	query := `
		SELECT uuid, password_hash 
		FROM users 
		WHERE email = $1
	`

	var userID uuid.UUID
	var passwordHash string

	err := adapter.db.QueryRow(query, email).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, "", ErrUserNotFound
		}
		return uuid.Nil, "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return userID, passwordHash, nil
}

// parseInsertError парсит ошибку INSERT и возвращает кастомную ошибку
func (adapter *PostgresAdapter) parseInsertError(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return nil // Не PostgreSQL ошибка
	}

	// Код 23505 = unique_violation
	if pqErr.Code != "23505" {
		return nil
	}

	// Определяем, какое поле вызвало конфликт, по имени constraint
	// Constraint обычно называется: <table>_<column>_key
	constraintName := strings.ToLower(pqErr.Constraint)

	switch {
	case strings.Contains(constraintName, "email"):
		return ErrEmailAlreadyExists
	case strings.Contains(constraintName, "username"), strings.Contains(constraintName, "user_name"):
		return ErrUsernameAlreadyExists
	default:
		// Если не смогли определить по имени constraint — смотрим в сообщение
		detail := strings.ToLower(pqErr.Detail)
		switch {
		case strings.Contains(detail, "email"):
			return ErrEmailAlreadyExists
		case strings.Contains(detail, "username"):
			return ErrUsernameAlreadyExists
		default:
			return fmt.Errorf("%w: unique violation on %s", ErrDatabaseError, pqErr.Constraint)
		}
	}
}

// HealthCheck проверяет подключение к БД
func (adapter *PostgresAdapter) HealthCheck() error {
	return adapter.db.Ping()
}

// Close закрывает соединение с БД
func (adapter *PostgresAdapter) Close() error {
	return adapter.db.Close()
}
