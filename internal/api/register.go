package api

import (
	"encoding/json"
	"net/http"
	"messenger/internal/auth"
)

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(registerResponse{Success: false, Error: "invalid request"})
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(registerResponse{Success: false, Error: "missing fields"})
		return
	}
	err := auth.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(registerResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(registerResponse{Success: true})
}
