package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"messenger/internal/db"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Карта: userID -> соединение
var Clients = make(map[int]*websocket.Conn)

type WSIncoming struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	ChatID    int    `json:"chat_id"`
	FileURL   string `json:"file_url,omitempty"`
	FileName  string `json:"file_name,omitempty"`
	MessageID int    `json:"message_id,omitempty"`
	Emoji     string `json:"emoji,omitempty"`
}

type WSOutgoing struct {
	Type      string                   `json:"type"`
	SenderID  int                      `json:"sender_id"`
	Username  string                   `json:"username,omitempty"`
	Text      string                   `json:"text,omitempty"`
	SentAt    string                   `json:"sent_at"`
	ChatID    int                      `json:"chat_id"`
	FileURL   string                   `json:"file_url,omitempty"`
	FileName  string                   `json:"file_name,omitempty"`
	MessageID int                      `json:"message_id,omitempty"`
	Emoji     string                   `json:"emoji,omitempty"`
	Reactions []map[string]interface{} `json:"reactions,omitempty"`
}

func HandleWS(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "No token", http.StatusUnauthorized)
		return
	}
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	userID, ok := claims["user_id"].(float64)
	if !ok {
		http.Error(w, "Invalid user_id", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS upgrade error:", err)
		return
	}
	Clients[int(userID)] = conn
	log.Printf("Пользователь %d подключился по WebSocket", int(userID))
	defer func() {
		conn.Close()
		delete(Clients, int(userID))
		log.Printf("Пользователь %d отключился", int(userID))
	}()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WS read error:", err)
			break
		}
		var in WSIncoming
		if err := json.Unmarshal(msg, &in); err != nil || in.ChatID == 0 {
			continue
		}

		// Валидация входных данных
		if in.Type == "message" && (len(in.Text) == 0 || len(in.Text) > 1000) {
			continue
		}

		// Базовая фильтрация HTML-тегов для защиты от XSS
		if in.Text != "" {
			in.Text = strings.ReplaceAll(in.Text, "<", "&lt;")
			in.Text = strings.ReplaceAll(in.Text, ">", "&gt;")
			in.Text = strings.ReplaceAll(in.Text, "\"", "&quot;")
			in.Text = strings.ReplaceAll(in.Text, "'", "&#39;")
		}

		sentAt := time.Now().UTC().Format(time.RFC3339)
		var username string
		_ = db.DB.QueryRow("SELECT username FROM users WHERE id = ?", int(userID)).Scan(&username)
		if in.Type == "file" && in.FileURL != "" {
			_, err = db.DB.Exec("INSERT INTO messages (chat_id, sender_id, text, sent_at) VALUES (?, ?, ?, ?)", in.ChatID, int(userID), "[file]"+in.FileURL, sentAt)
			if err != nil {
				log.Println("Ошибка сохранения file-сообщения:", err)
			}
			out := WSOutgoing{
				Type:     "file",
				SenderID: int(userID),
				Username: username,
				FileURL:  in.FileURL,
				FileName: in.FileName,
				SentAt:   sentAt,
				ChatID:   in.ChatID,
			}
			outMsg, _ := json.Marshal(out)
			SendToChatMembers(in.ChatID, outMsg)
			continue
		}
		if in.Type == "reaction" && in.MessageID > 0 && in.Emoji != "" {
			// Сохраняем реакцию
			_, err = db.DB.Exec(`INSERT INTO message_reactions (message_id, user_id, emoji, reacted_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
				ON CONFLICT(message_id, user_id) DO UPDATE SET emoji=excluded.emoji, reacted_at=CURRENT_TIMESTAMP`, in.MessageID, int(userID), in.Emoji)
			if err != nil {
				log.Println("Ошибка сохранения реакции:", err)
			}
			// Получаем все реакции к сообщению
			rows, _ := db.DB.Query(`SELECT user_id, emoji FROM message_reactions WHERE message_id = ?`, in.MessageID)
			var reactions []map[string]interface{}
			for rows.Next() {
				var uid int
				var emoji string
				if err := rows.Scan(&uid, &emoji); err == nil {
					reactions = append(reactions, map[string]interface{}{"user_id": uid, "emoji": emoji})
				}
			}
			rows.Close()
			out := WSOutgoing{
				Type:      "reaction",
				SenderID:  int(userID),
				MessageID: in.MessageID,
				Emoji:     in.Emoji,
				Reactions: reactions,
			}
			outMsg, _ := json.Marshal(out)
			SendToChatMembers(in.ChatID, outMsg)
			continue
		}
		if in.Type != "message" {
			continue
		}
		_, err = db.DB.Exec("INSERT INTO messages (chat_id, sender_id, text, sent_at) VALUES (?, ?, ?, ?)", in.ChatID, int(userID), in.Text, sentAt)
		if err != nil {
			log.Println("Ошибка сохранения сообщения:", err)
		}
		out := WSOutgoing{
			Type:     "message",
			SenderID: int(userID),
			Username: username,
			Text:     in.Text,
			SentAt:   sentAt,
			ChatID:   in.ChatID,
		}
		outMsg, _ := json.Marshal(out)
		SendToChatMembers(in.ChatID, outMsg)
	}
}

// Отправка сообщения участникам чата
func SendToChatMembers(chatID int, message []byte) {
	// Получаем список участников чата
	rows, err := db.DB.Query("SELECT user_id FROM chat_members WHERE chat_id = ?", chatID)
	if err != nil {
		log.Println("Ошибка получения участников чата:", err)
		return
	}
	defer rows.Close()

	var memberIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err == nil {
			memberIDs = append(memberIDs, userID)
		}
	}

	// Отправляем сообщение только участникам чата
	for _, userID := range memberIDs {
		if conn, exists := Clients[userID]; exists && conn != nil {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Ошибка отправки сообщения пользователю %d: %v", userID, err)
				// Удаляем неактивное соединение
				delete(Clients, userID)
			}
		}
	}
}
