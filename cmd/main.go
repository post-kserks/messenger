package main

import (
	"fmt"
	"net/http"

	"messenger/internal/db"
	"messenger/internal/api"
	"messenger/internal/auth"
)

func main() {
	if err := db.Init(); err != nil {
		panic("DB init error: " + err.Error())
	}
	if err := db.Ping(); err != nil {
		panic("DB ping error: " + err.Error())
	}

	http.HandleFunc("/register", api.RegisterHandler)
	http.HandleFunc("/login", api.LoginHandler)
	http.HandleFunc("/user", auth.AuthMiddleware(api.GetUserHandler))
	http.HandleFunc("/me", auth.AuthMiddleware(api.GetMeHandler))
	http.HandleFunc("/contacts", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.GetContactsHandler(w, r)
		} else if r.Method == http.MethodPost {
			api.AddContactHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/chat/private", auth.AuthMiddleware(api.CreatePrivateChatHandler))
	http.HandleFunc("/chat/group", auth.AuthMiddleware(api.CreateGroupChatHandler))
	http.HandleFunc("/chats", auth.AuthMiddleware(api.GetChatsHandler))
	http.HandleFunc("/message", auth.AuthMiddleware(api.SendMessageHandler))
	http.HandleFunc("/messages", auth.AuthMiddleware(api.GetMessagesHandler))

	// Статические файлы
	http.Handle("/", http.FileServer(http.Dir("static")))

	fmt.Println("Messenger server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}
