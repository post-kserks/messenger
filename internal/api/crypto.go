package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"messenger/internal/crypto"
)

// GetUserPublicKeyHandler получает публичный ключ пользователя
func GetUserPublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем user_id из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userIDStr := pathParts[3] // /api/users/{user_id}/public-key
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получаем публичный ключ
	userKey, err := crypto.GetUserPublicKey(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ключ пользователя не найден"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userKey)
}

// GetChatParticipantsKeysHandler получает ключи участников чата
func GetChatParticipantsKeysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем chat_id из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chatIDStr := pathParts[3] // /api/chats/{chat_id}/participants-keys
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получаем ключи участников чата
	participants, err := crypto.GetChatParticipantsKeys(chatID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка получения ключей участников"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participants)
}

// GetChatEncryptionStatusHandler получает статус шифрования для чата
func GetChatEncryptionStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем chat_id из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chatIDStr := pathParts[3] // /api/chats/{chat_id}/encryption-status
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получаем статус шифрования
	status, err := crypto.GetChatEncryptionStatus(chatID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка получения статуса шифрования"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// UpdateUserPublicKeyHandler обновляет публичный ключ пользователя
func UpdateUserPublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем user_id из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userIDStr := pathParts[3] // /api/users/{user_id}/public-key
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получаем данные из тела запроса
	var req struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный JSON"})
		return
	}

	// Сохраняем публичный ключ
	err = crypto.SaveUserPublicKey(userID, req.PublicKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ключ обновлен успешно"})
}

// MigrateUsersHandler генерирует ключи для существующих пользователей
func MigrateUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Проверяем авторизацию (только админ может выполнять миграцию)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Выполняем миграцию
	err := crypto.MigrateExistingUsers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка миграции: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Миграция завершена успешно"})
}
