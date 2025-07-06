package api

import (
	"bytes"
	"encoding/json"
	"messenger/internal/db"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRegisterLoginCreateChatAPI(t *testing.T) {
	os.Setenv("DB_PATH", ":memory:")
	db.Init()
	db.DB.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, email TEXT UNIQUE, password TEXT);`)
	db.DB.Exec(`CREATE TABLE chats (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, is_group BOOLEAN);`)
	db.DB.Exec(`CREATE TABLE chat_members (chat_id INTEGER, user_id INTEGER, PRIMARY KEY (chat_id, user_id));`)
	db.DB.Exec(`CREATE TABLE messages (id INTEGER PRIMARY KEY AUTOINCREMENT, chat_id INTEGER, sender_id INTEGER, text TEXT, sent_at TIMESTAMP);`)

	// Регистрация
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"username": "apiuser", "email": "api@ex.com", "password": "123"})
	r := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	Router(w, r)
	if w.Code != 200 {
		t.Fatalf("Регистрация через API не удалась: %d", w.Code)
	}

	// Логин
	w2 := httptest.NewRecorder()
	loginBody, _ := json.Marshal(map[string]string{"email": "api@ex.com", "password": "123"})
	r2 := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(loginBody))
	Router(w2, r2)
	if w2.Code != 200 {
		t.Fatalf("Логин через API не удался: %d", w2.Code)
	}

	// Создание чата
	w3 := httptest.NewRecorder()
	chatBody, _ := json.Marshal(map[string]interface{}{"name": "TestChat", "user_ids": []int{1}})
	r3 := httptest.NewRequest(http.MethodPost, "/api/chats", bytes.NewReader(chatBody))
	Router(w3, r3)
	if w3.Code != 201 {
		t.Fatalf("Создание чата через API не удалось: %d", w3.Code)
	}
}
