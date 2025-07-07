package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"messenger/internal/api"
	"messenger/internal/db"
	"messenger/internal/utils"
)

// Сервер с graceful shutdown
type Server struct {
	httpServer *http.Server
	wg         sync.WaitGroup
}

func NewServer(addr string) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	// Настройка маршрутов
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.HandleFunc("/", api.RouterOptimized) // Используем оптимизированный роутер

	log.Println("Сервер запущен на порту:", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Завершение работы сервера...")
	return s.httpServer.Shutdown(ctx)
}

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env не найден, используются переменные окружения по умолчанию")
	}

	// Инициализируем базу данных
	db.Init()

	// Запускаем очистку старых файлов
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	utils.CleanupOldFiles(uploadDir, 30*24*time.Hour) // 30 дней

	// Получаем порт
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Создаем сервер
	server := NewServer(":" + port)

	// Канал для сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ждем сигнала завершения
	<-stop

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при завершении сервера: %v", err)
	}

	log.Println("Сервер успешно завершен")
}
