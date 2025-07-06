package auth

import (
	"errors"
	"messenger/internal/db"
	"messenger/internal/models"
	"messenger/internal/utils"
)

// LoginUser проверяет логин и пароль, возвращает JWT-токен
func LoginUser(emailOrUsername, password string) (string, error) {
	var user models.User
	err := db.DB.Get(&user, "SELECT * FROM users WHERE email=$1 OR username=$1", emailOrUsername)
	if err != nil {
		return "", errors.New("user not found")
	}
	if !utils.CheckPassword(user.Password, password) {
		return "", errors.New("invalid password")
	}
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}
