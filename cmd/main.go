package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"messenger/internal/api"
	"messenger/internal/db"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env не найден, используются переменные окружения по умолчанию")
	}

	db.Init()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.HandleFunc("/", api.Router)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Сервер запущен на порту:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
