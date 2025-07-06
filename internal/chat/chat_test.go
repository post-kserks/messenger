package chat

import (
	"messenger/internal/db"
	"os"
	"testing"
)

func TestCreateChatAndMessage(t *testing.T) {
	os.Setenv("DB_PATH", ":memory:")
	db.Init()
	db.DB.Exec(`CREATE TABLE chats (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, is_group BOOLEAN);`)
	db.DB.Exec(`CREATE TABLE chat_members (chat_id INTEGER, user_id INTEGER, PRIMARY KEY (chat_id, user_id));`)
	db.DB.Exec(`CREATE TABLE messages (id INTEGER PRIMARY KEY AUTOINCREMENT, chat_id INTEGER, sender_id INTEGER, text TEXT, sent_at TIMESTAMP);`)

	// Создание чата
	res, err := db.DB.Exec("INSERT INTO chats (name, is_group) VALUES (?, ?)", "TestChat", false)
	if err != nil {
		t.Fatal(err)
	}
	chatID, _ := res.LastInsertId()
	_, err = db.DB.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES (?, ?)", chatID, 1)
	if err != nil {
		t.Fatal(err)
	}

	// Отправка сообщения
	_, err = db.DB.Exec("INSERT INTO messages (chat_id, sender_id, text) VALUES (?, ?, ?)", chatID, 1, "Hello!")
	if err != nil {
		t.Fatal(err)
	}
}
