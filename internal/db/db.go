package db

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

// Init инициализирует подключение к базе данных
func Init() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://golem:golem@localhost:5432/messenger?sslmode=disable"
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return fmt.Errorf("db connect error: %w", err)
	}
	DB = db
	return nil
}

// Ping проверяет соединение с базой данных
func Ping() error {
	if DB == nil {
		return fmt.Errorf("db not initialized")
	}
	return DB.Ping()
}
