package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Открываем базу данных
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "messenger.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Проверяем, существует ли таблица user_keys
	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='user_keys'").Scan(&tableExists)
	if err == nil && tableExists > 0 {
		log.Println("Миграция шифрования уже применена!")
		return
	}

	// Читаем SQL миграцию
	migrationSQL, err := os.ReadFile("migrations/002_encryption.sql")
	if err != nil {
		log.Fatalf("Ошибка чтения файла миграции: %v", err)
	}

	// Выполняем миграцию
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		log.Fatalf("Ошибка выполнения миграции: %v", err)
	}

	log.Println("Миграция шифрования успешно применена!")

	// Проверяем созданные таблицы
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name IN ('user_keys', 'encrypted_messages')")
	if err != nil {
		log.Printf("Ошибка проверки таблиц: %v", err)
		return
	}
	defer rows.Close()

	log.Println("Созданные таблицы шифрования:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err == nil {
			log.Printf("  - %s", tableName)
		}
	}
}
