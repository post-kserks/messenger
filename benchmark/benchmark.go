package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Результаты тестирования
type BenchmarkResult struct {
	TestName     string
	Duration     time.Duration
	Requests     int
	SuccessCount int
	ErrorCount   int
	AvgLatency   time.Duration
}

// Тест производительности
func runBenchmark(testName string, requests int, concurrency int, testFunc func() error) BenchmarkResult {
	start := time.Now()
	var wg sync.WaitGroup
	var successCount, errorCount int
	var mutex sync.Mutex

	// Создаем воркеры
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requests/concurrency; j++ {
				if err := testFunc(); err != nil {
					mutex.Lock()
					errorCount++
					mutex.Unlock()
				} else {
					mutex.Lock()
					successCount++
					mutex.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)
	avgLatency := duration / time.Duration(requests)

	return BenchmarkResult{
		TestName:     testName,
		Duration:     duration,
		Requests:     requests,
		SuccessCount: successCount,
		ErrorCount:   errorCount,
		AvgLatency:   avgLatency,
	}
}

// Тест получения сообщений
func testGetMessages() error {
	resp, err := http.Get("http://localhost:8080/api/messages?chat_id=1")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return nil
}

// Тест получения чатов
func testGetChats() error {
	resp, err := http.Get("http://localhost:8080/api/chats?user_id=1")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return nil
}

// Тест создания сообщения
func testCreateMessage() error {
	message := map[string]interface{}{
		"chat_id": 1,
		"text":    "Тестовое сообщение",
	}

	jsonData, _ := json.Marshal(message)
	resp, err := http.Post("http://localhost:8080/api/messages",
		"application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return nil
}

// Тест поиска пользователей
func testSearchUsers() error {
	resp, err := http.Get("http://localhost:8080/api/search_users?q=test")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return nil
}

func main() {
	fmt.Println("🚀 Запуск тестов производительности...")
	fmt.Println("Убедитесь, что сервер запущен на localhost:8080")
	fmt.Println()

	// Конфигурация тестов
	tests := []struct {
		name        string
		requests    int
		concurrency int
		testFunc    func() error
	}{
		{"Получение сообщений", 1000, 10, testGetMessages},
		{"Получение чатов", 500, 5, testGetChats},
		{"Создание сообщений", 200, 5, testCreateMessage},
		{"Поиск пользователей", 300, 5, testSearchUsers},
	}

	var results []BenchmarkResult

	// Запускаем тесты
	for _, test := range tests {
		fmt.Printf("Тестируем: %s...\n", test.name)
		result := runBenchmark(test.name, test.requests, test.concurrency, test.testFunc)
		results = append(results, result)

		fmt.Printf("✅ %s завершен за %v\n", test.name, result.Duration)
		fmt.Printf("   Запросов: %d, Успешно: %d, Ошибок: %d\n",
			result.Requests, result.SuccessCount, result.ErrorCount)
		fmt.Printf("   Средняя задержка: %v\n", result.AvgLatency)
		fmt.Printf("   RPS: %.2f\n\n",
			float64(result.SuccessCount)/result.Duration.Seconds())
	}

	// Сводка
	fmt.Println("📊 Сводка результатов:")
	fmt.Println("========================")

	totalRequests := 0
	totalSuccess := 0
	totalDuration := time.Duration(0)

	for _, result := range results {
		totalRequests += result.Requests
		totalSuccess += result.SuccessCount
		totalDuration += result.Duration

		fmt.Printf("%s:\n", result.TestName)
		fmt.Printf("  Время: %v\n", result.Duration)
		fmt.Printf("  Успешность: %.1f%%\n",
			float64(result.SuccessCount)/float64(result.Requests)*100)
		fmt.Printf("  RPS: %.2f\n",
			float64(result.SuccessCount)/result.Duration.Seconds())
		fmt.Printf("  Средняя задержка: %v\n\n", result.AvgLatency)
	}

	overallRPS := float64(totalSuccess) / totalDuration.Seconds()
	fmt.Printf("Общий RPS: %.2f\n", overallRPS)
	fmt.Printf("Общая успешность: %.1f%%\n",
		float64(totalSuccess)/float64(totalRequests)*100)
}
