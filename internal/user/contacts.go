package user

import (
	"errors"
	"messenger/internal/db"
	"messenger/internal/models"
)

// AddContact добавляет пользователя в контакты
func AddContact(userID, contactID int) error {
	if userID == contactID {
		return errors.New("cannot add yourself to contacts")
	}
	// Проверяем, существует ли пользователь
	var count int
	err := db.DB.Get(&count, "SELECT COUNT(*) FROM users WHERE id=$1", contactID)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("user not found")
	}
	// Проверяем, не добавлен ли уже в контакты
	err = db.DB.Get(&count, "SELECT COUNT(*) FROM contacts WHERE user_id=$1 AND contact_id=$2",
		userID, contactID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user already in contacts")
	}
	// Добавляем в контакты
	_, err = db.DB.Exec("INSERT INTO contacts (user_id, contact_id) VALUES ($1, $2)",
		userID, contactID)
	return err
}

// GetContacts получает список контактов пользователя
func GetContacts(userID int) ([]models.User, error) {
	var contacts []models.User
	err := db.DB.Select(&contacts, `
		SELECT u.id, u.username, u.email FROM users u
		JOIN contacts c ON u.id = c.contact_id
		WHERE c.user_id = $1
		ORDER BY u.username
	`, userID)
	return contacts, err
}
