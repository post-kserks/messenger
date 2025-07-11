package crypto

import (
	"database/sql"
	"fmt"
	"log"
	"messenger/internal/db"
	"messenger/internal/models"
)

// GetChatParticipantsKeys получает публичные ключи всех участников чата
func GetChatParticipantsKeys(chatID int) ([]models.ChatParticipantKey, error) {
	query := `
		SELECT cm.user_id, uk.public_key, u.username
		FROM chat_members cm
		JOIN user_keys uk ON cm.user_id = uk.user_id
		JOIN users u ON cm.user_id = u.id
		WHERE cm.chat_id = ?
		ORDER BY u.username
	`

	rows, err := db.DB.Query(query, chatID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса ключей участников: %w", err)
	}
	defer rows.Close()

	var participants []models.ChatParticipantKey
	for rows.Next() {
		var participant models.ChatParticipantKey
		err := rows.Scan(&participant.UserID, &participant.PublicKey, &participant.Username)
		if err != nil {
			log.Printf("Ошибка сканирования ключа участника: %v", err)
			continue
		}
		participants = append(participants, participant)
	}

	return participants, nil
}

// SaveUserPublicKey сохраняет публичный ключ пользователя
func SaveUserPublicKey(userID int, publicKey string) error {
	// Проверяем корректность ключа
	if err := ValidatePublicKey(publicKey); err != nil {
		return fmt.Errorf("неверный публичный ключ: %w", err)
	}

	// Проверяем, существует ли уже ключ для пользователя
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM user_keys WHERE user_id = ?)", userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования ключа: %w", err)
	}

	if exists {
		// Обновляем существующий ключ
		_, err = db.DB.Exec("UPDATE user_keys SET public_key = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?", publicKey, userID)
		if err != nil {
			return fmt.Errorf("ошибка обновления ключа: %w", err)
		}
	} else {
		// Создаем новый ключ
		_, err = db.DB.Exec("INSERT INTO user_keys (user_id, public_key) VALUES (?, ?)", userID, publicKey)
		if err != nil {
			return fmt.Errorf("ошибка создания ключа: %w", err)
		}
	}

	return nil
}

// GetUserPublicKey получает публичный ключ пользователя
func GetUserPublicKey(userID int) (*models.UserKey, error) {
	var userKey models.UserKey
	query := "SELECT user_id, public_key, created_at, updated_at FROM user_keys WHERE user_id = ?"

	err := db.DB.QueryRow(query, userID).Scan(&userKey.UserID, &userKey.PublicKey, &userKey.CreatedAt, &userKey.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ключ пользователя %d не найден", userID)
		}
		return nil, fmt.Errorf("ошибка получения ключа: %w", err)
	}

	return &userKey, nil
}

// DeleteUserKey удаляет ключ пользователя
func DeleteUserKey(userID int) error {
	_, err := db.DB.Exec("DELETE FROM user_keys WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("ошибка удаления ключа: %w", err)
	}
	return nil
}

// CheckUserHasKey проверяет, есть ли у пользователя ключ
func CheckUserHasKey(userID int) (bool, error) {
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM user_keys WHERE user_id = ?)", userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки ключа: %w", err)
	}
	return exists, nil
}

// GetUsersWithoutKeys получает список пользователей без ключей
func GetUsersWithoutKeys() ([]int, error) {
	query := `
		SELECT u.id
		FROM users u
		LEFT JOIN user_keys uk ON u.id = uk.user_id
		WHERE uk.user_id IS NULL
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса пользователей без ключей: %w", err)
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err == nil {
			userIDs = append(userIDs, userID)
		}
	}

	return userIDs, nil
}

// GetChatEncryptionStatus получает статус шифрования для чата
func GetChatEncryptionStatus(chatID int) (map[string]interface{}, error) {
	// Получаем всех участников чата
	participants, err := GetChatParticipantsKeys(chatID)
	if err != nil {
		return nil, err
	}

	// Получаем общее количество участников
	var totalMembers int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM chat_members WHERE chat_id = ?", chatID).Scan(&totalMembers)
	if err != nil {
		return nil, fmt.Errorf("ошибка подсчета участников: %w", err)
	}

	// Подсчитываем участников с ключами
	membersWithKeys := len(participants)
	encryptionEnabled := membersWithKeys == totalMembers && totalMembers > 0

	return map[string]interface{}{
		"chat_id":               chatID,
		"total_members":         totalMembers,
		"members_with_keys":     membersWithKeys,
		"encryption_enabled":    encryptionEnabled,
		"encryption_percentage": float64(membersWithKeys) / float64(totalMembers) * 100,
	}, nil
}

// MigrateExistingUsers генерирует ключи для существующих пользователей без ключей
func MigrateExistingUsers() error {
	usersWithoutKeys, err := GetUsersWithoutKeys()
	if err != nil {
		return fmt.Errorf("ошибка получения пользователей без ключей: %w", err)
	}

	log.Printf("Найдено %d пользователей без ключей", len(usersWithoutKeys))

	for _, userID := range usersWithoutKeys {
		// Генерируем новую пару ключей
		keyPair, err := GenerateKeyPair()
		if err != nil {
			log.Printf("Ошибка генерации ключей для пользователя %d: %v", userID, err)
			continue
		}

		// Сохраняем публичный ключ
		err = SaveUserPublicKey(userID, keyPair.PublicKey)
		if err != nil {
			log.Printf("Ошибка сохранения ключа для пользователя %d: %v", userID, err)
			continue
		}

		log.Printf("Сгенерирован ключ для пользователя %d", userID)
	}

	return nil
}
