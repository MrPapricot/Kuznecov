package repository

import (
	db_adapter "db_adapter/src"
	"fmt"
	"time"

	"errors"
	user_models "shared/user_data"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	NoUserFound           = errors.New("No user found for given token")
	TokenExpired          = errors.New("Token expired")
	InvalidToken          = errors.New("Invalid token")
	DatabaseError         = errors.New("Database error")
	UsernameAlreadyExists = errors.New("Username already exist")
	PasswordMissmath      = errors.New("Invalid password")
)

type UserRepository struct {
	db              *db_adapter.PostgresAdapter
	jwt_secret      string
	jwt_expire_time time.Duration
}

func NewUserRepository(db *db_adapter.PostgresAdapter, jwt_secret string, jwt_expire_time time.Duration) UserRepository {
	return UserRepository{db: db, jwt_secret: jwt_secret, jwt_expire_time: jwt_expire_time}
}

func (repository *UserRepository) get_uuid_from_token(token_string string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(token_string, &user_models.TokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(repository.jwt_secret), nil
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
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	return claims.UUID, nil
}

func (repository *UserRepository) ChangeUsername(token string, new_username string) (string, string, error) {
	user_uuid, err := repository.get_uuid_from_token(token)
	if err != nil {
		return "", "", err
	}

	old_username, err := repository.db.UpdateUsername(user_uuid, new_username)
	if err != nil {
		switch {
		case errors.Is(err, db_adapter.ErrUserNotFound):
			return "", "", NoUserFound
		case errors.Is(err, db_adapter.ErrUsernameAlreadyExists):
			return "", "", UsernameAlreadyExists
		case errors.Is(err, db_adapter.ErrDatabaseError):
			return "", "", DatabaseError
		default:
			return "", "", err
		}
	}

	return old_username, new_username, nil
}

func (repository *UserRepository) ChangePassword(token string, old_password string, new_password string) error {
	user_uuid, err := repository.get_uuid_from_token(token)
	if err != nil {
		return err
	}

	// 1. Получаем текущий хеш пароля из БД
	current_hash, err := repository.db.GetPasswordHash(user_uuid)
	if err != nil {
		switch {
		case errors.Is(err, db_adapter.ErrUserNotFound):
			return NoUserFound
		case errors.Is(err, db_adapter.ErrDatabaseError):
			return DatabaseError
		default:
			return err
		}
	}

	// 2. Проверяем старый пароль
	is_valid, err := argon2id.ComparePasswordAndHash(old_password, current_hash)
	if err != nil {
		return fmt.Errorf("failed to compare password: %w", err)
	}
	if !is_valid {
		return PasswordMissmath
	}

	// 3. Хешируем новый пароль
	new_hash, err := argon2id.CreateHash(new_password, argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 4. Обновляем пароль в БД
	err = repository.db.UpdatePasswordHash(user_uuid, new_hash)
	if err != nil {
		switch {
		case errors.Is(err, db_adapter.ErrUserNotFound):
			return NoUserFound
		case errors.Is(err, db_adapter.ErrDatabaseError):
			return DatabaseError
		default:
			return err
		}
	}

	return nil
}

func (repository *UserRepository) DeleteUser(token string, password string) error {
	user_uuid, err := repository.get_uuid_from_token(token)
	if err != nil {
		return err
	}

	// 1. Получаем текущий хеш пароля из БД
	current_hash, err := repository.db.GetPasswordHash(user_uuid)
	if err != nil {
		switch {
		case errors.Is(err, db_adapter.ErrUserNotFound):
			return NoUserFound
		case errors.Is(err, db_adapter.ErrDatabaseError):
			return DatabaseError
		default:
			return err
		}
	}

	// 2. Проверяем старый пароль
	is_valid, err := argon2id.ComparePasswordAndHash(password, current_hash)
	if err != nil {
		return fmt.Errorf("failed to compare password: %w", err)
	}
	if !is_valid {
		return PasswordMissmath
	}

	err = repository.db.DeleteUser(user_uuid)
	if err != nil {
		switch {
		case errors.Is(err, db_adapter.ErrUserNotFound):
			return NoUserFound
		case errors.Is(err, db_adapter.ErrTransactionError):
			fallthrough
		case errors.Is(err, db_adapter.ErrDatabaseError):
			return DatabaseError
		default:
			return err
		}
	}

	return nil
}

func (repository *UserRepository) GetUserInfo(token string) (db_adapter.UserInfo, error) {
	user_uuid, err := repository.get_uuid_from_token(token)
	if err != nil {
		return db_adapter.UserInfo{}, err
	}

	user_info, err := repository.db.GetUserInfo(user_uuid)
	if err != nil {
		switch {
		case errors.Is(err, db_adapter.ErrUserNotFound):
			return db_adapter.UserInfo{}, NoUserFound
		case errors.Is(err, db_adapter.ErrDatabaseError):
			return db_adapter.UserInfo{}, DatabaseError
		default:
			return db_adapter.UserInfo{}, err
		}
	}

	return user_info, nil
}
