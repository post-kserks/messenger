package auth

import (
	"bytes"
	"encoding/json"
	"messenger/internal/db"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRegisterAndLogin(t *testing.T) {
	os.Setenv("DB_PATH", ":memory:")
	db.Init()
	db.DB.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, email TEXT UNIQUE, password TEXT)`)

	// Регистрация
	w := httptest.NewRecorder()
	body, _ := json.Marshal(RegisterRequest{Username: "testuser", Email: "test@example.com", Password: "123456"})
	r := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	RegisterHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("Регистрация не удалась: %d", w.Code)
	}

	// Повторная регистрация (конфликт)
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	RegisterHandler(w2, r2)
	if w2.Code != 409 {
		t.Error("Ожидался конфликт при повторной регистрации")
	}

	// Логин
	w3 := httptest.NewRecorder()
	loginBody, _ := json.Marshal(LoginRequest{Email: "test@example.com", Password: "123456"})
	r3 := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(loginBody))
	LoginHandler(w3, r3)
	if w3.Code != 200 {
		t.Fatalf("Логин не удался: %d", w3.Code)
	}
}
