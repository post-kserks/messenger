package auth

import (
	"errors"
	"messenger/internal/db"
	"messenger/internal/utils"
)

// RegisterUser регистрирует нового пользователя
func RegisterUser(username, email, password string) error {
	// Проверка уникальности email
	var count int
	err := db.DB.Get(&count, "SELECT COUNT(*) FROM users WHERE email=$1", email)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("email already in use")
	}
	// Проверка уникальности username
	err = db.DB.Get(&count, "SELECT COUNT(*) FROM users WHERE username=$1", username)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("username already in use")
	}
	// Хешируем пароль
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	// Сохраняем пользователя
	_, err = db.DB.Exec(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		username, email, hash,
	)
	return err
}
