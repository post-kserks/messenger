package models

import "time"

// Message представляет сообщение в чате
type Message struct {
    ID       int       `db:"id" json:"id"`
    ChatID   int       `db:"chat_id" json:"chat_id"`
    SenderID int       `db:"sender_id" json:"sender_id"`
    Text     string    `db:"text" json:"text"`
    SentAt   time.Time `db:"sent_at" json:"sent_at"`
}
