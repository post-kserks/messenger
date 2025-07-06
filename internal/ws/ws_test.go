package ws

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"messenger/internal/db"

	"github.com/gorilla/websocket"
)

// Тестовый обработчик без проверки JWT
type testWS struct{}

func (testWS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func TestWebSocketEcho(t *testing.T) {
	os.Setenv("DB_PATH", ":memory:")
	db.Init()
	db.DB.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, email TEXT, password TEXT);`)
	db.DB.Exec(`INSERT INTO users (id, username, email, password) VALUES (1, 'wsuser', 'ws@ex.com', 'hash')`)

	ts := httptest.NewServer(testWS{})
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws"

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("WebSocket dial error: %v", err)
	}
	defer c.Close()

	type msg struct {
		Type   string `json:"type"`
		Text   string `json:"text"`
		ChatID int    `json:"chat_id"`
	}
	c.WriteJSON(msg{Type: "message", Text: "hello", ChatID: 1})
	c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, resp, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("WebSocket read error: %v", err)
	}
	if string(resp) == "" {
		t.Error("Пустой ответ от WebSocket")
	}
}
