package api

import (
	"encoding/json"
	"messenger/internal/auth"
	"messenger/internal/db"
	"messenger/internal/ws"
	"net/http"

	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Router(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	switch {
	case r.URL.Path == "/ws":
		ws.HandleWS(w, r)
	case r.URL.Path == "/api/register":
		auth.RegisterHandler(w, r)
	case r.URL.Path == "/api/login":
		auth.LoginHandler(w, r)
	case r.URL.Path == "/api/messages":
		if r.Method == http.MethodGet {
			getMessagesHandler(w, r)
		} else if r.Method == http.MethodPost {
			createMessageHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case r.URL.Path == "/api/chats":
		if r.Method == http.MethodPost {
			createChatHandler(w, r)
		} else if r.Method == http.MethodGet {
			getChatsHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case r.URL.Path == "/api/contacts":
		if r.Method == http.MethodPost {
			addContactHandler(w, r)
		} else if r.Method == http.MethodGet {
			getContactsHandler(w, r)
		} else if r.Method == http.MethodDelete {
			deleteContactHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case r.URL.Path == "/api/user":
		if r.Method == http.MethodGet {
			getUserHandler(w, r)
		} else if r.Method == http.MethodPut {
			updateUserHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case r.URL.Path == "/api/search":
		searchUsersHandler(w, r)
	case r.URL.Path == "/api/upload":
		uploadFileHandler(w, r)
	case r.URL.Path == "/api/reaction":
		if r.Method == http.MethodPost {
			addReactionHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case r.URL.Path == "/api/reactions":
		getReactionsHandler(w, r)
	case r.URL.Path == "/api/unread":
		getUnreadHandler(w, r)
	case r.URL.Path == "/api/mark_read":
		if r.Method == http.MethodPost {
			markReadHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case r.URL.Path == "/":
		http.ServeFile(w, r, "static/index.html")
	default:
		http.NotFound(w, r)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("WebSocket endpoint (заглушка)"))
}

// Получение истории сообщений (MVP: chat_id=1)
func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Ошибка: Метод не поддерживается, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		chatID = "1"
	}
	rows, err := db.DB.Query("SELECT m.id, m.sender_id, u.username, m.text, m.sent_at FROM messages m JOIN users u ON m.sender_id = u.id WHERE m.chat_id = ? ORDER BY m.sent_at ASC LIMIT 100", chatID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	defer rows.Close()
	var msgs []map[string]interface{}
	for rows.Next() {
		var id, sender int
		var username, text, sentAt string
		if err := rows.Scan(&id, &sender, &username, &text, &sentAt); err == nil {
			msg := map[string]interface{}{
				"id":        id,
				"sender_id": sender,
				"username":  username,
				"text":      text,
				"sent_at":   sentAt,
			}
			if len(text) > 6 && text[:6] == "[file]" {
				msg["type"] = "file"
				msg["file_url"] = text[6:]
				msg["text"] = ""
			}
			msgs = append(msgs, msg)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

// Создание нового сообщения (POST)
func createMessageHandler(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		ChatID int    `json:"chat_id"`
		Text   string `json:"text"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ChatID == 0 || req.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректные данные запроса, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}

	// Валидация длины сообщения
	if len(req.Text) > 1000 {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Сообщение слишком длинное, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	// Базовая фильтрация HTML-тегов (защита от XSS)
	req.Text = strings.ReplaceAll(req.Text, "<", "&lt;")
	req.Text = strings.ReplaceAll(req.Text, ">", "&gt;")
	req.Text = strings.ReplaceAll(req.Text, "\"", "&quot;")
	req.Text = strings.ReplaceAll(req.Text, "'", "&#39;")

	// Получаем sender_id из токена авторизации
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Нет токена авторизации, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Невалидный токен, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	uidFloat, ok := claims["user_id"].(float64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: user_id не найден в токене, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	senderID := int(uidFloat)

	// Проверяем, что пользователь является участником чата
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id = ? AND user_id = ?)", req.ChatID, senderID).Scan(&exists)
	if err != nil || !exists {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Ошибка: Пользователь не является участником чата, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	// Создаем сообщение
	res, err := db.DB.Exec("INSERT INTO messages (chat_id, sender_id, text, sent_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)", req.ChatID, senderID, req.Text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}

	messageID, _ := res.LastInsertId()

	// Получаем username отправителя
	var username string
	err = db.DB.QueryRow("SELECT username FROM users WHERE id = ?", senderID).Scan(&username)
	if err != nil {
		log.Printf("Ошибка получения username: %v", err)
		username = "Unknown"
	}

	// Отправляем уведомление через WebSocket
	wsMessage := map[string]interface{}{
		"type":      "new_message",
		"id":        messageID,
		"chat_id":   req.ChatID,
		"sender_id": senderID,
		"username":  username,
		"text":      req.Text,
		"sent_at":   time.Now().UTC().Format(time.RFC3339),
	}

	wsMessageBytes, _ := json.Marshal(wsMessage)
	ws.SendToChatMembers(req.ChatID, wsMessageBytes)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        messageID,
		"chat_id":   req.ChatID,
		"sender_id": senderID,
		"text":      req.Text,
	})
}

// Создание чата (POST)
func createChatHandler(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		Name    string `json:"name"`
		UserIDs []int  `json:"user_ids"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || len(req.UserIDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректные данные запроса, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	res, err := db.DB.Exec("INSERT INTO chats (name, is_group) VALUES (?, ?)", req.Name, len(req.UserIDs) > 1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	chatID, _ := res.LastInsertId()
	for _, uid := range req.UserIDs {
		_, _ = db.DB.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES (?, ?)", chatID, uid)
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"chat_id": chatID})
}

// Получение чатов пользователя (GET)
func getChatsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректный запрос, отсутствует параметр user_id, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	rows, err := db.DB.Query(`SELECT c.id, c.name, c.is_group FROM chats c JOIN chat_members m ON c.id = m.chat_id WHERE m.user_id = ?`, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	defer rows.Close()
	var chats []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var isGroup bool
		if err := rows.Scan(&id, &name, &isGroup); err == nil {
			chat := map[string]interface{}{
				"id": id, "name": name, "is_group": isGroup,
			}
			// Получаем последнее сообщение
			var lastMsgText, lastMsgTime string
			_ = db.DB.QueryRow("SELECT text, sent_at FROM messages WHERE chat_id = ? ORDER BY sent_at DESC LIMIT 1", id).Scan(&lastMsgText, &lastMsgTime)
			chat["last_msg_text"] = lastMsgText
			chat["last_msg_time"] = lastMsgTime
			// Получаем количество непрочитанных сообщений
			var unreadCount int
			_ = db.DB.QueryRow(`SELECT COUNT(*) FROM messages m LEFT JOIN chat_reads cr ON cr.user_id = ? AND cr.chat_id = m.chat_id WHERE m.chat_id = ? AND m.sender_id != ? AND (cr.last_read IS NULL OR m.sent_at > cr.last_read)`, userID, id, userID).Scan(&unreadCount)
			chat["unread_count"] = unreadCount
			chats = append(chats, chat)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

// Добавление контакта по никнейму (POST)
func addContactHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Нет токена авторизации, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Невалидный токен, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	currentUserID := int(claims["user_id"].(float64))

	type reqT struct {
		Username string `json:"username"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректные данные запроса, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}

	// Находим пользователя по никнейму или username
	var contactID int
	err = db.DB.QueryRow("SELECT id FROM users WHERE (nickname = ? OR username = ?) AND id != ?", req.Username, req.Username, currentUserID).Scan(&contactID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("Ошибка: Пользователь с никнеймом/именем %s не найден, запрос: %s %s", req.Username, r.Method, r.URL.Path)
		return
	}

	// Проверяем, не добавлен ли уже этот пользователь в контакты
	var exists int
	err = db.DB.QueryRow("SELECT 1 FROM contacts WHERE user_id = ? AND contact_id = ?", currentUserID, contactID).Scan(&exists)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		log.Printf("Ошибка: Пользователь уже в контактах, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	// Добавляем в контакты
	_, err = db.DB.Exec("INSERT INTO contacts (user_id, contact_id) VALUES (?, ?)", currentUserID, contactID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Контакт добавлен",
		"contact_id": contactID,
	})
}

// Получение списка контактов (GET)
func getContactsHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Нет токена авторизации, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Невалидный токен, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	currentUserID := int(claims["user_id"].(float64))

	rows, err := db.DB.Query(`SELECT u.id, u.username, u.email, u.nickname FROM users u JOIN contacts c ON u.id = c.contact_id WHERE c.user_id = ?`, currentUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	defer rows.Close()
	var contacts []map[string]interface{}
	for rows.Next() {
		var id int
		var username, email, nickname string
		if err := rows.Scan(&id, &username, &email, &nickname); err == nil {
			contacts = append(contacts, map[string]interface{}{
				"id": id, "username": username, "email": email, "nickname": nickname,
			})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contacts)
}

// Удаление контакта (DELETE)
func deleteContactHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Нет токена авторизации, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Невалидный токен, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	currentUserID := int(claims["user_id"].(float64))

	type reqT struct {
		ContactID int `json:"contact_id"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ContactID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректные данные запроса, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	_, err = db.DB.Exec("DELETE FROM contacts WHERE user_id = ? AND contact_id = ?", currentUserID, req.ContactID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Контакт удален",
	})
}

// Получение информации о пользователе (GET)
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("Ошибка: Нет токена авторизации, запрос: %s %s", r.Method, r.URL.Path)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("Ошибка: Невалидный токен, запрос: %s %s", r.Method, r.URL.Path)
			return
		}
		uidFloat, ok := claims["user_id"].(float64)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("Ошибка: user_id не найден в токене, запрос: %s %s", r.Method, r.URL.Path)
			return
		}
		id = strconv.Itoa(int(uidFloat))
	}
	row := db.DB.QueryRow("SELECT id, username, email, nickname FROM users WHERE id = ?", id)
	var uid int
	var username, email, nickname string
	if err := row.Scan(&uid, &username, &email, &nickname); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": uid, "username": username, "email": email, "nickname": nickname,
	})
}

// Обновление профиля пользователя (PUT)
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректные данные запроса, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}

	// Валидация никнейма
	if req.Nickname != "" {
		if len(req.Nickname) < 3 || len(req.Nickname) > 32 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Никнейм должен быть от 3 до 32 символов",
			})
			return
		}
		// Проверка символов
		for _, c := range req.Nickname {
			if !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && !(c >= '0' && c <= '9') && c != '_' {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Никнейм может содержать только буквы, цифры и _",
				})
				return
			}
		}
		// Проверка уникальности (исключая текущего пользователя)
		var exists int
		err := db.DB.QueryRow("SELECT 1 FROM users WHERE nickname = ? AND id != ?", req.Nickname, req.ID).Scan(&exists)
		if err == nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Никнейм уже занят",
			})
			return
		}
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Ошибка: Не удалось зашифровать пароль, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
			return
		}
		_, err = db.DB.Exec("UPDATE users SET username=?, email=?, nickname=?, password=? WHERE id=?", req.Username, req.Email, req.Nickname, string(hash), req.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
			return
		}
	} else {
		_, err := db.DB.Exec("UPDATE users SET username=?, email=?, nickname=? WHERE id=?", req.Username, req.Email, req.Nickname, req.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Профиль обновлен",
	})
}

// Поиск пользователей по никнейму
func searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Нет токена авторизации, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Невалидный токен, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	currentUserID := int(claims["user_id"].(float64))

	q := r.URL.Query().Get("q")
	if q == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректный запрос, отсутствует параметр q, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	// Ищем пользователей по никнейму или username, исключая текущего пользователя
	rows, err := db.DB.Query(`
		SELECT u.id, u.username, u.email, u.nickname,
		       CASE WHEN c.contact_id IS NOT NULL THEN 1 ELSE 0 END as is_contact
		FROM users u
		LEFT JOIN contacts c ON u.id = c.contact_id AND c.user_id = ?
		WHERE (u.nickname LIKE ? OR u.username LIKE ?) AND u.id != ?
		LIMIT 20
	`, currentUserID, "%"+q+"%", "%"+q+"%", currentUserID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var username, email, nickname string
		var isContact int
		if err := rows.Scan(&id, &username, &email, &nickname, &isContact); err == nil {
			users = append(users, map[string]interface{}{
				"id":         id,
				"username":   username,
				"email":      email,
				"nickname":   nickname,
				"is_contact": isContact == 1,
			})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Загрузка файла (POST)
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Ошибка: Метод не поддерживается, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	// Проверяем авторизацию
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Нет токена авторизации, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: Невалидный токен, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	uidFloat, ok := claims["user_id"].(float64)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Ошибка: user_id не найден в токене, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	senderID := int(uidFloat)

	// Получаем chat_id из формы
	err = r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Не удалось распарсить форму, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}

	chatIDStr := r.FormValue("chat_id")
	if chatIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Отсутствует chat_id, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil || chatID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректный chat_id, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	// Проверяем, что пользователь является участником чата
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id = ? AND user_id = ?)", chatID, senderID).Scan(&exists)
	if err != nil || !exists {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Ошибка: Пользователь не является участником чата, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Не удалось получить файл из запроса, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	defer file.Close()

	// Проверяем размер файла (максимум 10MB)
	if handler.Size > 10<<20 {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Файл слишком большой, запрос: %s %s", r.Method, r.URL.Path)
		return
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	os.MkdirAll(uploadDir, 0755)
	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + filepath.Ext(handler.Filename)
	fpath := filepath.Join(uploadDir, filename)
	out, err := os.Create(fpath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось создать файл, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось скопировать файл, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}

	fileURL := "/uploads/" + filename
	sentAt := time.Now().UTC().Format(time.RFC3339)

	// Создаем сообщение с файлом в базе данных
	res, err := db.DB.Exec("INSERT INTO messages (chat_id, sender_id, text, sent_at) VALUES (?, ?, ?, ?)", chatID, senderID, "[file]"+fileURL, sentAt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось сохранить сообщение с файлом, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}

	messageID, _ := res.LastInsertId()

	// Получаем username отправителя
	var username string
	err = db.DB.QueryRow("SELECT username FROM users WHERE id = ?", senderID).Scan(&username)
	if err != nil {
		log.Printf("Ошибка получения username: %v", err)
		username = "Unknown"
	}

	// Отправляем уведомление через WebSocket
	wsMessage := map[string]interface{}{
		"type":      "new_file",
		"id":        messageID,
		"chat_id":   chatID,
		"sender_id": senderID,
		"username":  username,
		"file_url":  fileURL,
		"file_name": handler.Filename,
		"sent_at":   sentAt,
	}

	wsMessageBytes, _ := json.Marshal(wsMessage)
	ws.SendToChatMembers(chatID, wsMessageBytes)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        messageID,
		"chat_id":   chatID,
		"sender_id": senderID,
		"file_url":  fileURL,
		"file_name": handler.Filename,
	})
}

// Добавить/обновить реакцию
func addReactionHandler(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		MessageID int    `json:"message_id"`
		UserID    int    `json:"user_id"`
		Emoji     string `json:"emoji"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MessageID == 0 || req.UserID == 0 || req.Emoji == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректные данные запроса, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	_, err := db.DB.Exec(`INSERT INTO message_reactions (message_id, user_id, emoji, reacted_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(message_id, user_id) DO UPDATE SET emoji=excluded.emoji, reacted_at=CURRENT_TIMESTAMP`, req.MessageID, req.UserID, req.Emoji)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Получить реакции к сообщению
func getReactionsHandler(w http.ResponseWriter, r *http.Request) {
	msgID := r.URL.Query().Get("message_id")
	if msgID == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректный запрос, отсутствует параметр message_id, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	rows, err := db.DB.Query(`SELECT user_id, emoji FROM message_reactions WHERE message_id = ?`, msgID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	defer rows.Close()
	var reactions []map[string]interface{}
	for rows.Next() {
		var uid int
		var emoji string
		if err := rows.Scan(&uid, &emoji); err == nil {
			reactions = append(reactions, map[string]interface{}{
				"user_id": uid, "emoji": emoji,
			})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reactions)
}

// Получение количества непрочитанных сообщений по чатам
func getUnreadHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректный запрос, отсутствует параметр user_id, запрос: %s %s", r.Method, r.URL.Path)
		return
	}
	rows, err := db.DB.Query(`SELECT m.chat_id, COUNT(*) FROM messages m JOIN chat_members cm ON m.chat_id = cm.chat_id LEFT JOIN chat_reads cr ON cr.user_id = cm.user_id AND cr.chat_id = m.chat_id WHERE cm.user_id = ? AND m.sender_id != ? AND (cr.last_read IS NULL OR m.sent_at > cr.last_read) GROUP BY m.chat_id`, userID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	defer rows.Close()
	var unread []map[string]interface{}
	for rows.Next() {
		var chatID, count int
		if err := rows.Scan(&chatID, &count); err == nil {
			unread = append(unread, map[string]interface{}{"chat_id": chatID, "count": count})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unread)
}

// Отметить сообщения в чате как прочитанные
func markReadHandler(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		UserID int `json:"user_id"`
		ChatID int `json:"chat_id"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == 0 || req.ChatID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Ошибка: Некорректные данные запроса, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	// Получаем время последнего сообщения в чате
	var lastMsgTime string
	err := db.DB.QueryRow("SELECT MAX(sent_at) FROM messages WHERE chat_id = ?", req.ChatID).Scan(&lastMsgTime)
	if lastMsgTime == "" {
		// Если сообщений нет, не обновляем last_read
		w.WriteHeader(http.StatusOK)
		return
	}
	_, err = db.DB.Exec(`INSERT INTO chat_reads (user_id, chat_id, last_read) VALUES (?, ?, ?)
		ON CONFLICT(user_id, chat_id) DO UPDATE SET last_read=?`, req.UserID, req.ChatID, lastMsgTime, lastMsgTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Ошибка: Не удалось выполнить запрос к базе данных, запрос: %s %s, ошибка: %v", r.Method, r.URL.Path, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
