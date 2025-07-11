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
	ID            int       `json:"id"`
	ChatID        int       `json:"chat_id"`
	SenderID      int       `json:"sender_id"`
	Text          string    `json:"text,omitempty"`           // Обычное сообщение
	EncryptedData string    `json:"encrypted_data,omitempty"` // Зашифрованное сообщение
	Nonce         string    `json:"nonce,omitempty"`          // Для зашифрованных сообщений
	IsEncrypted   bool      `json:"is_encrypted"`             // Флаг шифрования
	SentAt        time.Time `json:"sent_at"`
}

type EncryptedMessage struct {
	ID            int       `json:"id"`
	ChatID        int       `json:"chat_id"`
	SenderID      int       `json:"sender_id"`
	EncryptedData string    `json:"encrypted_data"`
	Nonce         string    `json:"nonce"`
	SentAt        time.Time `json:"sent_at"`
}

type UserKey struct {
	UserID    int    `json:"user_id"`
	PublicKey string `json:"public_key"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ChatParticipantKey struct {
	UserID    int    `json:"user_id"`
	PublicKey string `json:"public_key"`
	Username  string `json:"username,omitempty"`
}

type Contact struct {
	UserID    int `json:"user_id"`
	ContactID int `json:"contact_id"`
}

type ChatMember struct {
	ChatID int `json:"chat_id"`
	UserID int `json:"user_id"`
}
