package chat

import (
	"errors"
	"messenger/internal/db"
	"messenger/internal/models"
)

// CreatePrivateChat создает личный чат между двумя пользователями
func CreatePrivateChat(userID1, userID2 int) (*models.Chat, error) {
	// Проверяем, существует ли уже чат между этими пользователями
	var existingChat models.Chat
	err := db.DB.Get(&existingChat, `
		SELECT c.* FROM chats c
		JOIN chat_members cm1 ON c.id = cm1.chat_id
		JOIN chat_members cm2 ON c.id = cm2.chat_id
		WHERE c.is_group = false
		AND cm1.user_id = $1 AND cm2.user_id = $2
	`, userID1, userID2)
	if err == nil {
		return &existingChat, nil // Чат уже существует
	}
	// Создаем новый чат
	var chatID int
	err = db.DB.QueryRow(
		"INSERT INTO chats (name, is_group) VALUES ($1, $2) RETURNING id",
		"", false,
	).Scan(&chatID)
	if err != nil {
		return nil, err
	}
	// Добавляем участников
	_, err = db.DB.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2), ($1, $3)",
		chatID, userID1, userID2)
	if err != nil {
		return nil, err
	}
	return &models.Chat{ID: chatID, Name: "", IsGroup: false}, nil
}

// CreateGroupChat создает групповой чат
func CreateGroupChat(name string, creatorID int, memberIDs []int) (*models.Chat, error) {
	if name == "" {
		return nil, errors.New("group name is required")
	}
	// Создаем чат
	var chatID int
	err := db.DB.QueryRow(
		"INSERT INTO chats (name, is_group) VALUES ($1, $2) RETURNING id",
		name, true,
	).Scan(&chatID)
	if err != nil {
		return nil, err
	}
	// Добавляем создателя и участников
	memberIDs = append([]int{creatorID}, memberIDs...)
	for _, memberID := range memberIDs {
		_, err = db.DB.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)",
			chatID, memberID)
		if err != nil {
			return nil, err
		}
	}
	return &models.Chat{ID: chatID, Name: name, IsGroup: true}, nil
}

// GetUserChats получает все чаты пользователя
func GetUserChats(userID int) ([]models.Chat, error) {
	var chats []models.Chat
	err := db.DB.Select(&chats, `
		SELECT c.* FROM chats c
		JOIN chat_members cm ON c.id = cm.chat_id
		WHERE cm.user_id = $1
		ORDER BY c.id DESC
	`, userID)
	return chats, err
}
