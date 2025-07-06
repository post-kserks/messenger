package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword хеширует пароль
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckPassword сравнивает пароль и хеш
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
