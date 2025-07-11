package crypto

import (
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Ошибка генерации ключей: %v", err)
	}

	if keyPair.PublicKey == "" {
		t.Error("Публичный ключ пустой")
	}

	if keyPair.PrivateKey == "" {
		t.Error("Приватный ключ пустой")
	}

	// Проверяем, что ключи разные
	if keyPair.PublicKey == keyPair.PrivateKey {
		t.Error("Публичный и приватный ключи одинаковые")
	}

	// Проверяем валидность ключей
	if err := keyPair.IsValid(); err != nil {
		t.Errorf("Сгенерированная пара ключей невалидна: %v", err)
	}
}

func TestEncryptDecryptMessage(t *testing.T) {
	// Генерируем пару ключей
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Ошибка генерации ключей: %v", err)
	}

	// Тестовое сообщение
	originalMessage := "Привет, это тестовое сообщение! 🔐"

	// Шифруем сообщение
	encrypted, err := EncryptMessage(originalMessage, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("Ошибка шифрования: %v", err)
	}

	if encrypted.EncryptedData == "" {
		t.Error("Зашифрованные данные пустые")
	}

	if encrypted.Nonce == "" {
		t.Error("Nonce пустой")
	}

	// Расшифровываем сообщение
	decrypted, err := DecryptMessage(encrypted, keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("Ошибка расшифровки: %v", err)
	}

	// Проверяем, что расшифрованное сообщение совпадает с оригиналом
	if decrypted != originalMessage {
		t.Errorf("Расшифрованное сообщение не совпадает с оригиналом. Ожидалось: %s, Получено: %s", originalMessage, decrypted)
	}
}

func TestEncryptDecryptMultipleMessages(t *testing.T) {
	// Генерируем пару ключей
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Ошибка генерации ключей: %v", err)
	}

	messages := []string{
		"Первое сообщение",
		"Второе сообщение с эмодзи 😊",
		"Третье сообщение с кириллицей и цифрами 123",
		"",
		"Очень длинное сообщение " + string(make([]byte, 1000)), // 1000 символов
	}

	for i, originalMessage := range messages {
		// Шифруем сообщение
		encrypted, err := EncryptMessage(originalMessage, keyPair.PublicKey)
		if err != nil {
			t.Fatalf("Ошибка шифрования сообщения %d: %v", i, err)
		}

		// Расшифровываем сообщение
		decrypted, err := DecryptMessage(encrypted, keyPair.PrivateKey)
		if err != nil {
			t.Fatalf("Ошибка расшифровки сообщения %d: %v", i, err)
		}

		// Проверяем совпадение
		if decrypted != originalMessage {
			t.Errorf("Сообщение %d не совпадает. Ожидалось: %s, Получено: %s", i, originalMessage, decrypted)
		}
	}
}

func TestValidatePublicKey(t *testing.T) {
	// Генерируем валидный ключ
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Ошибка генерации ключей: %v", err)
	}

	// Тестируем валидный ключ
	if err := ValidatePublicKey(keyPair.PublicKey); err != nil {
		t.Errorf("Валидный ключ не прошел проверку: %v", err)
	}

	// Тестируем невалидные ключи
	invalidKeys := []string{
		"",
		"invalid-base64!@#",
		"short",
		"very-long-key-that-is-not-base64-encoded-properly-and-contains-invalid-characters",
	}

	for _, invalidKey := range invalidKeys {
		if err := ValidatePublicKey(invalidKey); err == nil {
			t.Errorf("Невалидный ключ '%s' прошел проверку", invalidKey)
		}
	}
}

func TestValidatePrivateKey(t *testing.T) {
	// Генерируем валидный ключ
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Ошибка генерации ключей: %v", err)
	}

	// Тестируем валидный ключ
	if err := ValidatePrivateKey(keyPair.PrivateKey); err != nil {
		t.Errorf("Валидный ключ не прошел проверку: %v", err)
	}

	// Тестируем невалидные ключи
	invalidKeys := []string{
		"",
		"invalid-base64!@#",
		"short",
		"very-long-key-that-is-not-base64-encoded-properly-and-contains-invalid-characters",
	}

	for _, invalidKey := range invalidKeys {
		if err := ValidatePrivateKey(invalidKey); err == nil {
			t.Errorf("Невалидный ключ '%s' прошел проверку", invalidKey)
		}
	}
}

func TestSanitizeKeyPair(t *testing.T) {
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Ошибка генерации ключей: %v", err)
	}

	sanitized := keyPair.SanitizeForClient()

	// Проверяем, что публичный ключ присутствует
	if sanitized["public_key"] != keyPair.PublicKey {
		t.Error("Публичный ключ не совпадает")
	}

	// Проверяем, что приватный ключ отсутствует
	if _, exists := sanitized["private_key"]; exists {
		t.Error("Приватный ключ присутствует в очищенной структуре")
	}
}

func TestEncryptWithWrongKey(t *testing.T) {
	// Генерируем две пары ключей
	keyPair1, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Ошибка генерации первой пары ключей: %v", err)
	}

	keyPair2, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Ошибка генерации второй пары ключей: %v", err)
	}

	// Шифруем сообщение ключом первого пользователя
	originalMessage := "Секретное сообщение"
	encrypted, err := EncryptMessage(originalMessage, keyPair1.PublicKey)
	if err != nil {
		t.Fatalf("Ошибка шифрования: %v", err)
	}

	// Пытаемся расшифровать ключом второго пользователя
	_, err = DecryptMessage(encrypted, keyPair2.PrivateKey)
	if err == nil {
		t.Error("Сообщение расшифровалось чужим ключом")
	}
}

func BenchmarkGenerateKeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateKeyPair()
		if err != nil {
			b.Fatalf("Ошибка генерации ключей: %v", err)
		}
	}
}

func BenchmarkEncryptDecrypt(b *testing.B) {
	keyPair, err := GenerateKeyPair()
	if err != nil {
		b.Fatalf("Ошибка генерации ключей: %v", err)
	}

	message := "Тестовое сообщение для бенчмарка"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encrypted, err := EncryptMessage(message, keyPair.PublicKey)
		if err != nil {
			b.Fatalf("Ошибка шифрования: %v", err)
		}

		_, err = DecryptMessage(encrypted, keyPair.PrivateKey)
		if err != nil {
			b.Fatalf("Ошибка расшифровки: %v", err)
		}
	}
}
