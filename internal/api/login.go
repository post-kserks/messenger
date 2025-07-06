package api

import (
	"encoding/json"
	"net/http"
	"messenger/internal/auth"
)

type loginRequest struct {
	EmailOrUsername string `json:"email_or_username"`
	Password        string `json:"password"`
}

type loginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(loginResponse{Success: false, Error: "invalid request"})
		return
	}
	if req.EmailOrUsername == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(loginResponse{Success: false, Error: "missing fields"})
		return
	}
	token, err := auth.LoginUser(req.EmailOrUsername, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(loginResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(loginResponse{Success: true, Token: token})
}
