package user

import (
	"messenger/internal/db"
	"os"
	"testing"
)

func TestAddAndGetContact(t *testing.T) {
	os.Setenv("DB_PATH", ":memory:")
	db.Init()
	db.DB.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, email TEXT, password TEXT);`)
	db.DB.Exec(`CREATE TABLE contacts (user_id INTEGER, contact_id INTEGER, PRIMARY KEY (user_id, contact_id));`)

	// Добавление пользователей
	res1, _ := db.DB.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", "user1", "u1@mail.com", "p1")
	res2, _ := db.DB.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", "user2", "u2@mail.com", "p2")
	id1, _ := res1.LastInsertId()
	id2, _ := res2.LastInsertId()

	// Добавление контакта
	_, err := db.DB.Exec("INSERT INTO contacts (user_id, contact_id) VALUES (?, ?)", id1, id2)
	if err != nil {
		t.Fatal(err)
	}

	// Получение контактов
	rows, err := db.DB.Query("SELECT contact_id FROM contacts WHERE user_id = ?", id1)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for rows.Next() {
		var cid int64
		rows.Scan(&cid)
		if cid == id2 {
			found = true
		}
	}
	if !found {
		t.Error("Контакт не найден")
	}
}
