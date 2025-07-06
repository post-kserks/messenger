package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"messenger/internal/auth"
	"messenger/internal/chat"
	"messenger/internal/models"
)

type sendMessageRequest struct {
	ChatID int    `json:"chat_id"`
	Text   string `json:"text"`
}

type messageResponse struct {
	Success  bool              `json:"success"`
	Message  *models.Message   `json:"message,omitempty"`
	Messages []models.Message  `json:"messages,omitempty"`
	Error    string            `json:"error,omitempty"`
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(messageResponse{Success: false, Error: "invalid request"})
		return
	}
	message, err := chat.SendMessage(req.ChatID, userID, req.Text)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(messageResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(messageResponse{Success: true, Message: message})
}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(messageResponse{Success: false, Error: "missing chat_id"})
		return
	}
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(messageResponse{Success: false, Error: "invalid chat_id"})
		return
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	messages, err := chat.GetChatMessages(chatID, userID, limit)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(messageResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(messageResponse{Success: true, Messages: messages})
}
