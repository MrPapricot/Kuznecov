package repository

import (
	db_adapter "db_adapter/src"
	"errors"
	"fmt"

	user_models "shared/user_data"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	InvalidToken  = errors.New("invalid token")
	TokenExpired  = errors.New("token expired")
	NoUserFound   = errors.New("no user found for given token")
	DatabaseError = errors.New("database error")
)

type StatsRepository struct {
	db        *db_adapter.PostgresAdapter
	jwtSecret string
}

func NewStatsRepository(db *db_adapter.PostgresAdapter, jwtSecret string) StatsRepository {
	return StatsRepository{db: db, jwtSecret: jwtSecret}
}

func (r *StatsRepository) GetGlobalStats() (db_adapter.GlobalStats, error) {
	return r.db.GetGlobalStats()
}

func (r *StatsRepository) GetUserStats(token string) (db_adapter.UserStats, error) {
	// Валидация токена и проверка существования пользователя
	userUUID, err := r.getAndValidateUUIDFromToken(token)
	if err != nil {
		return db_adapter.UserStats{}, err
	}

	return r.db.GetUserStats(userUUID)
}

func (r *StatsRepository) getAndValidateUUIDFromToken(tokenString string) (uuid.UUID, error) {
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

	exists, err := r.db.UserExists(claims.UUID)
	if err != nil {
		return uuid.Nil, DatabaseError
	}
	if !exists {
		return uuid.Nil, NoUserFound
	}

	return claims.UUID, nil
}

func (r *StatsRepository) HealthCheck() error {
	return r.db.HealthCheck()
}
