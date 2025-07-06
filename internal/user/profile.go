package user

import (
	"errors"
	"messenger/internal/db"
	"messenger/internal/models"
)

// GetUserByID получает пользователя по ID
func GetUserByID(userID int) (*models.User, error) {
	var user models.User
	err := db.DB.Get(&user, "SELECT id, username, email FROM users WHERE id=$1", userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
