package db_adapter

import "github.com/google/uuid"

type UserBody struct {
	Email        string
	PasswordHash string
}

type DBAdapter interface {
	CreateUser(UserBody) (uuid.UUID, error)
	AuthUser(UserBody) (uuid.UUID, error)
	CheckToken(string) (bool, error)
}
