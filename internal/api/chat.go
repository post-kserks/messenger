package api

import (
	"encoding/json"
	"net/http"
	"messenger/internal/auth"
	"messenger/internal/chat"
	"messenger/internal/models"
)

type createPrivateChatRequest struct {
	UserID2 int `json:"user_id_2"`
}

type createGroupChatRequest struct {
	Name      string `json:"name"`
	MemberIDs []int  `json:"member_ids"`
}

type chatResponse struct {
	Success bool           `json:"success"`
	Chat    *models.Chat   `json:"chat,omitempty"`
	Chats   []models.Chat  `json:"chats,omitempty"`
	Error   string         `json:"error,omitempty"`
}

func CreatePrivateChatHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	var req createPrivateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(chatResponse{Success: false, Error: "invalid request"})
		return
	}
	chat, err := chat.CreatePrivateChat(userID, req.UserID2)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(chatResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(chatResponse{Success: true, Chat: chat})
}

func CreateGroupChatHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	var req createGroupChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(chatResponse{Success: false, Error: "invalid request"})
		return
	}
	chat, err := chat.CreateGroupChat(req.Name, userID, req.MemberIDs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(chatResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(chatResponse{Success: true, Chat: chat})
}

func GetChatsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	chats, err := chat.GetUserChats(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(chatResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(chatResponse{Success: true, Chats: chats})
}
