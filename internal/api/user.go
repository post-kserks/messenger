package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"messenger/internal/user"
	"messenger/internal/models"
	"messenger/internal/auth"
)

type userResponse struct {
	Success bool         `json:"success"`
	User    *models.User `json:"user,omitempty"`
	Error   string       `json:"error,omitempty"`
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(userResponse{Success: false, Error: "missing user id"})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(userResponse{Success: false, Error: "invalid user id"})
		return
	}
	user, err := user.GetUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(userResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(userResponse{Success: true, User: user})
}

func GetMeHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	user, err := user.GetUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(userResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(userResponse{Success: true, User: user})
}
