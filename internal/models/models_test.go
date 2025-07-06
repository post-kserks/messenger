package models

import (
	"testing"
	"time"
)

func TestModelStructs(t *testing.T) {
	u := User{ID: 1, Username: "u", Email: "e", Password: "p"}
	c := Chat{ID: 1, Name: "c", IsGroup: false}
	m := Message{ID: 1, ChatID: 1, SenderID: 1, Text: "hi", SentAt: time.Now()}
	cm := ChatMember{ChatID: 1, UserID: 1}
	ct := Contact{UserID: 1, ContactID: 2}
	if u.ID != 1 || c.ID != 1 || m.ID != 1 || cm.ChatID != 1 || ct.ContactID != 2 {
		t.Error("Ошибка в моделях")
	}
}
