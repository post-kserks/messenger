package api

import (
	"encoding/json"
	"net/http"
	"messenger/internal/auth"
	"messenger/internal/user"
	"messenger/internal/models"
)

type addContactRequest struct {
	ContactID int `json:"contact_id"`
}

type contactResponse struct {
	Success  bool         `json:"success"`
	Contacts []models.User `json:"contacts,omitempty"`
	Error    string       `json:"error,omitempty"`
}

func AddContactHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	var req addContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(contactResponse{Success: false, Error: "invalid request"})
		return
	}
	err := user.AddContact(userID, req.ContactID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(contactResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(contactResponse{Success: true})
}

func GetContactsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)
	contacts, err := user.GetContacts(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(contactResponse{Success: false, Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(contactResponse{Success: true, Contacts: contacts})
}
