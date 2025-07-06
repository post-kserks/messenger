package models

import "time"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Chat struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IsGroup bool   `json:"is_group"`
}

type Message struct {
	ID       int       `json:"id"`
	ChatID   int       `json:"chat_id"`
	SenderID int       `json:"sender_id"`
	Text     string    `json:"text"`
	SentAt   time.Time `json:"sent_at"`
}

type Contact struct {
	UserID    int `json:"user_id"`
	ContactID int `json:"contact_id"`
}

type ChatMember struct {
	ChatID int `json:"chat_id"`
	UserID int `json:"user_id"`
}
