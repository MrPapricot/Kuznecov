package db_adapter

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserInfo struct {
	UserName  string
	Email     string
	CreatedAt time.Time
}

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

func (adapter *PostgresAdapter) GetUserInfo(uuid uuid.UUID) (UserInfo, error) {
	query := `
		SELECT username, email, created_at
		FROM users
		WHERE uuid = $1
	`
	var user_info UserInfo

	err := adapter.db.QueryRow(query, uuid).Scan(&user_info.UserName, &user_info.Email, &user_info.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserInfo{}, ErrUserNotFound
		}
		fmt.Printf("%#v\n", err)
		return UserInfo{}, ErrDatabaseError
	}

	fmt.Printf("User_Info: %#v\n", user_info)

	return user_info, nil
}

func (adapter *PostgresAdapter) GetPasswordHash(userID uuid.UUID) (string, error) {
	query := `SELECT password_hash FROM users WHERE uuid = $1`
	var passwordHash string
	err := adapter.db.QueryRow(query, userID).Scan(&passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return passwordHash, nil
}

func (adapter *PostgresAdapter) UpdatePasswordHash(userID uuid.UUID, password_hash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE uuid = $2`
	result, err := adapter.db.Exec(query, password_hash, userID)
	if err != nil {
		fmt.Printf("Error: %#v\n", err)
		return ErrDatabaseError
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Printf("Error: %#v\n", err)
		return ErrDatabaseError
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (adapter *PostgresAdapter) UpdateUsername(userID uuid.UUID, newUsername string) (string, error) {
	// Сначала получаем текущее имя
	var oldUsername string
	err := adapter.db.QueryRow("SELECT username FROM users WHERE uuid = $1", userID).Scan(&oldUsername)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Обновляем имя
	query := `UPDATE users SET username = $1 WHERE uuid = $2`
	result, err := adapter.db.Exec(query, newUsername, userID)
	if err != nil {
		// Обрабатываем ошибку уникальности
		if parsedErr := adapter.parseUpdateError(err); parsedErr != nil {
			return "", parsedErr
		}
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if rowsAffected == 0 {
		return "", ErrUserNotFound
	}

	return oldUsername, nil
}

func (adapter *PostgresAdapter) DeleteUser(userID uuid.UUID) error {
	tx, err := adapter.db.Begin()
	if err != nil {
		fmt.Printf("Error starting transaction: %#v\n", err)
		return ErrTransactionError
	}
	defer tx.Rollback()

	// 1. Проверяем существование с блокировкой строки
	checkUserExistsQuery := `
		SELECT EXISTS(SELECT 1 FROM users WHERE uuid = $1)
	`
	var exists bool
	err = tx.QueryRow(checkUserExistsQuery, userID).Scan(&exists)
	if err != nil {
		fmt.Printf("Failed to find user: %#v\n", err)
		return ErrDatabaseError
	}
	if !exists {
		return ErrUserNotFound
	}

	// 2. Удаляем связи
	deleteRelationsQuery := `
		DELETE FROM rooms_users_relations
		WHERE member_uuid = $1
	`
	_, err = tx.Exec(deleteRelationsQuery, userID)
	if err != nil {
		fmt.Printf("Failed to delete user relations: %#v\n", err)
		return ErrDatabaseError
	}

	// 3. Удаляем пользователя
	deleteUserQuery := `
		DELETE FROM users
		WHERE uuid = $1
	`
	_, err = tx.Exec(deleteUserQuery, userID)
	if err != nil {
		return ErrDatabaseError
	}

	// 4. Коммитим транзакцию
	return tx.Commit()
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

func (adapter *PostgresAdapter) parseUpdateError(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return nil
	}

	if pqErr.Code != "23505" {
		return nil
	}

	constraintName := strings.ToLower(pqErr.Constraint)
	if strings.Contains(constraintName, "username") || strings.Contains(constraintName, "user_name") {
		return ErrUsernameAlreadyExists
	}

	detail := strings.ToLower(pqErr.Detail)
	if strings.Contains(detail, "username") {
		return ErrUsernameAlreadyExists
	}

	return fmt.Errorf("%w: unique violation on %s", ErrDatabaseError, pqErr.Constraint)
}

// HealthCheck проверяет подключение к БД
func (adapter *PostgresAdapter) HealthCheck() error {
	return adapter.db.Ping()
}

// Close закрывает соединение с БД
func (adapter *PostgresAdapter) Close() error {
	return adapter.db.Close()
}
