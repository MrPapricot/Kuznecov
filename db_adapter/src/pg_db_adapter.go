package db_adapter

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

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
	connect_url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", options.UserName, options.UserPassword, options.Host, options.Port, options.DBName)
	db, err := sqlx.Connect("postgres", connect_url)
	if err != nil {
		return PostgresAdapter{nil}, err
	}
	return PostgresAdapter{db}, nil
}

func (adapter *PostgresAdapter) CreateUser(UserBody) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (adapter *PostgresAdapter) AuthUser(UserBody) (uuid.UUID, error) {
	return uuid.New(), nil
}
