package chat

import (
	"errors"
	"messenger/internal/db"
	"messenger/internal/models"
)

// SendMessage отправляет сообщение в чат
func SendMessage(chatID, senderID int, text string) (*models.Message, error) {
	if text == "" {
		return nil, errors.New("message text cannot be empty")
	}
	// Проверяем, является ли отправитель участником чата
	var count int
	err := db.DB.Get(&count, "SELECT COUNT(*) FROM chat_members WHERE chat_id=$1 AND user_id=$2",
		chatID, senderID)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("user is not a member of this chat")
	}
	// Сохраняем сообщение
	var messageID int
	err = db.DB.QueryRow(
		"INSERT INTO messages (chat_id, sender_id, text) VALUES ($1, $2, $3) RETURNING id",
		chatID, senderID, text,
	).Scan(&messageID)
	if err != nil {
		return nil, err
	}
	// Получаем созданное сообщение
	var message models.Message
	err = db.DB.Get(&message, "SELECT * FROM messages WHERE id=$1", messageID)
	return &message, err
}

// GetChatMessages получает сообщения из чата
func GetChatMessages(chatID, userID int, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 50 // По умолчанию 50 сообщений
	}
	// Проверяем, является ли пользователь участником чата
	var count int
	err := db.DB.Get(&count, "SELECT COUNT(*) FROM chat_members WHERE chat_id=$1 AND user_id=$2",
		chatID, userID)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("user is not a member of this chat")
	}
	// Получаем сообщения
	var messages []models.Message
	err = db.DB.Select(&messages, `
		SELECT * FROM messages
		WHERE chat_id = $1
		ORDER BY sent_at DESC
		LIMIT $2
	`, chatID, limit)
	return messages, err
}
