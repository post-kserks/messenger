package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "messenger.db"
	}
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatalf("Ошибка пинга базы данных: %v", err)
	}
	log.Println("База данных подключена:", dbPath)
}

func GetDB() *sql.DB {
	return DB
}
