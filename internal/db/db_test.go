package db

import (
	"os"
	"testing"
)

func TestInitAndPing(t *testing.T) {
	os.Setenv("DB_PATH", ":memory:")
	Init()
	if DB == nil {
		t.Fatal("DB не инициализирована")
	}
	if err := DB.Ping(); err != nil {
		t.Fatalf("Ping не удался: %v", err)
	}
}
