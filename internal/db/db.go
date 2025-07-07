package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite3"
	}
	var dsn string
	if driver == "sqlite3" {
		dsn = os.Getenv("DB_PATH")
		if dsn == "" {
			dsn = "messenger.db"
		}
	} else if driver == "postgres" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")
		if sslmode == "" {
			sslmode = "disable"
		}
		dsn = "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode
	} else {
		log.Fatalf("Неизвестный драйвер базы данных: %s", driver)
	}
	var err error
	DB, err = sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatalf("Ошибка пинга базы данных: %v", err)
	}
	log.Println("База данных подключена через драйвер:", driver)
}

func GetDB() *sql.DB {
	return DB
}
