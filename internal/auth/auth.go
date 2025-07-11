package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"messenger/internal/crypto"
	"messenger/internal/db"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type RegisterResponse struct {
	Success  bool   `json:"success"`
	Token    string `json:"token,omitempty"`
	UserID   int    `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Error    string `json:"error,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success  bool   `json:"success"`
	Token    string `json:"token,omitempty"`
	UserID   int    `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Error    string `json:"error,omitempty"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Некорректный JSON"})
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" || req.Nickname == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Все поля обязательны"})
		return
	}
	// Проверка длины nickname
	if len(req.Nickname) < 3 || len(req.Nickname) > 32 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Никнейм должен быть от 3 до 32 символов"})
		return
	}
	// Проверка allowed chars
	for _, c := range req.Nickname {
		if !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && !(c >= '0' && c <= '9') && c != '_' {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Никнейм может содержать только буквы, цифры и _"})
			return
		}
	}
	// Проверка уникальности nickname
	var exists int
	err := db.DB.QueryRow("SELECT 1 FROM users WHERE nickname = ?", req.Nickname).Scan(&exists)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Никнейм уже занят"})
		return
	}
	// Проверка длины полей
	if len(req.Username) < 3 || len(req.Username) > 50 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Имя пользователя должно быть от 3 до 50 символов"})
		return
	}

	if len(req.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Пароль должен содержать минимум 6 символов"})
		return
	}

	if !strings.Contains(req.Email, "@") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Некорректный email"})
		return
	}

	// Базовая фильтрация HTML-тегов
	req.Username = strings.ReplaceAll(req.Username, "<", "&lt;")
	req.Username = strings.ReplaceAll(req.Username, ">", "&gt;")
	req.Username = strings.ReplaceAll(req.Username, "\"", "&quot;")
	req.Username = strings.ReplaceAll(req.Username, "'", "&#39;")

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Ошибка сервера"})
		return
	}
	result, err := db.DB.Exec("INSERT INTO users (username, email, password, nickname) VALUES (?, ?, ?, ?)", req.Username, req.Email, string(hash), req.Nickname)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Email или никнейм уже зарегистрирован"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Ошибка базы данных"})
		return
	}

	// Получаем ID созданного пользователя
	userID, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Ошибка получения ID пользователя"})
		return
	}

	// Генерируем пару ключей для нового пользователя
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Ошибка генерации ключей"})
		return
	}

	// Сохраняем публичный ключ в базу данных
	_, err = db.DB.Exec("INSERT INTO user_keys (user_id, public_key) VALUES (?, ?)", userID, keyPair.PublicKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Ошибка сохранения ключей"})
		return
	}

	// Генерируем JWT токен
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret_key"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Ошибка генерации токена"})
		return
	}

	json.NewEncoder(w).Encode(RegisterResponse{
		Success:  true,
		Token:    tokenString,
		UserID:   int(userID),
		Username: req.Username,
		Nickname: req.Nickname,
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Success: false, Error: "Некорректный JSON"})
		return
	}
	var id int
	var hash string
	var username, nickname string
	row := db.DB.QueryRow("SELECT id, password, username, nickname FROM users WHERE email = ?", req.Email)
	err := row.Scan(&id, &hash, &username, &nickname)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{Success: false, Error: "Пользователь не найден"})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Success: false, Error: "Ошибка базы данных"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{Success: false, Error: "Неверный пароль"})
		return
	}
	// Генерация JWT
	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Success: false, Error: "Ошибка генерации токена"})
		return
	}
	json.NewEncoder(w).Encode(LoginResponse{Success: true, Token: tokenString, UserID: id, Username: username, Nickname: nickname})
}
