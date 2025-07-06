package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"errors"
)

var jwtSecret = []byte("supersecretkey") // Лучше вынести в env

// GenerateJWT создает JWT-токен для пользователя
func GenerateJWT(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT проверяет JWT-токен и возвращает user_id
func ValidateJWT(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int(claims["user_id"].(float64))
		return userID, nil
	}
	return 0, errors.New("invalid token")
}
