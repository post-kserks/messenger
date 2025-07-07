package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Очередь для обработки файлов
var fileQueue = make(chan FileTask, 100)

type FileTask struct {
	File      multipart.File
	Header    *multipart.FileHeader
	UploadDir string
	Callback  func(string, error)
}

// Результат обработки файла
type FileResult struct {
	Filename string
	FileURL  string
	Size     int64
	Error    error
}

// Обработчик файлов
func FileProcessor() {
	for task := range fileQueue {
		go processFile(task)
	}
}

func processFile(task FileTask) {
	defer task.File.Close()

	// Создаем уникальное имя файла
	hash := md5.New()
	io.Copy(hash, task.File)
	task.File.Seek(0, 0) // Возвращаемся в начало файла

	fileHash := fmt.Sprintf("%x", hash.Sum(nil))
	ext := filepath.Ext(task.Header.Filename)
	filename := fileHash + ext

	// Создаем директорию если не существует
	if err := os.MkdirAll(task.UploadDir, 0755); err != nil {
		task.Callback("", err)
		return
	}

	filePath := filepath.Join(task.UploadDir, filename)

	// Создаем файл
	dst, err := os.Create(filePath)
	if err != nil {
		task.Callback("", err)
		return
	}
	defer dst.Close()

	// Копируем содержимое
	size, err := io.Copy(dst, task.File)
	if err != nil {
		task.Callback("", err)
		return
	}

	// Проверяем размер файла
	if size > 10*1024*1024 { // 10MB лимит
		os.Remove(filePath)
		task.Callback("", fmt.Errorf("файл слишком большой"))
		return
	}

	// Проверяем тип файла
	if !isAllowedFileType(ext) {
		os.Remove(filePath)
		task.Callback("", fmt.Errorf("неподдерживаемый тип файла"))
		return
	}

	fileURL := "/uploads/" + filename
	task.Callback(fileURL, nil)
}

// Проверка разрешенных типов файлов
func isAllowedFileType(ext string) bool {
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
		".txt":  true,
		".doc":  true,
		".docx": true,
		".mp3":  true,
		".mp4":  true,
		".avi":  true,
		".zip":  true,
		".rar":  true,
	}
	return allowedTypes[strings.ToLower(ext)]
}

// Асинхронная загрузка файла
func UploadFileAsync(file multipart.File, header *multipart.FileHeader, uploadDir string) chan FileResult {
	resultChan := make(chan FileResult, 1)

	task := FileTask{
		File:      file,
		Header:    header,
		UploadDir: uploadDir,
		Callback: func(fileURL string, err error) {
			if err != nil {
				resultChan <- FileResult{Error: err}
			} else {
				resultChan <- FileResult{
					Filename: header.Filename,
					FileURL:  fileURL,
					Size:     header.Size,
				}
			}
			close(resultChan)
		},
	}

	// Отправляем задачу в очередь
	select {
	case fileQueue <- task:
		// Задача отправлена успешно
	default:
		// Очередь переполнена, обрабатываем синхронно
		log.Println("Очередь файлов переполнена, обработка синхронно")
		go processFile(task)
	}

	return resultChan
}

// Очистка старых файлов
func CleanupOldFiles(uploadDir string, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Проверяем раз в день
		defer ticker.Stop()

		for range ticker.C {
			cleanupFiles(uploadDir, maxAge)
		}
	}()
}

func cleanupFiles(uploadDir string, maxAge time.Duration) {
	files, err := os.ReadDir(uploadDir)
	if err != nil {
		log.Printf("Ошибка чтения директории %s: %v", uploadDir, err)
		return
	}

	cutoff := time.Now().Add(-maxAge)
	var wg sync.WaitGroup

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		wg.Add(1)
		go func(filename string) {
			defer wg.Done()

			filePath := filepath.Join(uploadDir, filename)
			info, err := os.Stat(filePath)
			if err != nil {
				return
			}

			if info.ModTime().Before(cutoff) {
				os.Remove(filePath)
				log.Printf("Удален старый файл: %s", filename)
			}
		}(file.Name())
	}

	wg.Wait()
}

// Инициализация обработчика файлов
func init() {
	go FileProcessor()
}
