package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/nacl/box"
)

const (
	PublicKeySize = 32
	SecretKeySize = 32
	NonceSize     = 24
)

// KeyPair представляет пару ключей пользователя
type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// EncryptedMessage представляет зашифрованное сообщение
type EncryptedMessage struct {
	EncryptedData string `json:"encrypted_data"`
	Nonce         string `json:"nonce"`
	SenderID      int    `json:"sender_id"`
}

// GenerateKeyPair генерирует новую пару ключей для пользователя
func GenerateKeyPair() (*KeyPair, error) {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации ключей: %w", err)
	}

	return &KeyPair{
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey[:]),
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey[:]),
	}, nil
}

// EncryptMessage шифрует сообщение для получателя
func EncryptMessage(message string, recipientPublicKey string) (*EncryptedMessage, error) {
	// Декодируем публичный ключ получателя
	publicKeyBytes, err := base64.StdEncoding.DecodeString(recipientPublicKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования публичного ключа: %w", err)
	}

	if len(publicKeyBytes) != PublicKeySize {
		return nil, errors.New("неверный размер публичного ключа")
	}

	var publicKey [PublicKeySize]byte
	copy(publicKey[:], publicKeyBytes)

	// Генерируем временную пару ключей для этого сообщения
	ephemeralPublicKey, ephemeralPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации временных ключей: %w", err)
	}

	// Генерируем nonce
	var nonce [NonceSize]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("ошибка генерации nonce: %w", err)
	}

	// Шифруем сообщение
	encrypted := box.Seal(nil, []byte(message), &nonce, &publicKey, ephemeralPrivateKey)

	// Объединяем временный публичный ключ с зашифрованными данными
	combinedData := append(ephemeralPublicKey[:], encrypted...)

	return &EncryptedMessage{
		EncryptedData: base64.StdEncoding.EncodeToString(combinedData),
		Nonce:         base64.StdEncoding.EncodeToString(nonce[:]),
	}, nil
}

// DecryptMessage расшифровывает сообщение
func DecryptMessage(encryptedMessage *EncryptedMessage, privateKey string) (string, error) {
	// Декодируем приватный ключ
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования приватного ключа: %w", err)
	}

	if len(privateKeyBytes) != SecretKeySize {
		return "", errors.New("неверный размер приватного ключа")
	}

	var secretKey [SecretKeySize]byte
	copy(secretKey[:], privateKeyBytes)

	// Декодируем зашифрованные данные
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedMessage.EncryptedData)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования зашифрованных данных: %w", err)
	}

	// Декодируем nonce
	nonceBytes, err := base64.StdEncoding.DecodeString(encryptedMessage.Nonce)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования nonce: %w", err)
	}

	if len(nonceBytes) != NonceSize {
		return "", errors.New("неверный размер nonce")
	}

	var nonce [NonceSize]byte
	copy(nonce[:], nonceBytes)

	// Извлекаем временный публичный ключ из начала данных
	if len(encryptedData) < PublicKeySize {
		return "", errors.New("неверный размер зашифрованных данных")
	}

	var ephemeralPublicKey [PublicKeySize]byte
	copy(ephemeralPublicKey[:], encryptedData[:PublicKeySize])

	// Извлекаем зашифрованные данные
	actualEncryptedData := encryptedData[PublicKeySize:]

	// Расшифровываем сообщение
	decrypted, ok := box.Open(nil, actualEncryptedData, &nonce, &ephemeralPublicKey, &secretKey)
	if !ok {
		return "", errors.New("не удалось расшифровать сообщение")
	}

	return string(decrypted), nil
}

// ValidatePublicKey проверяет корректность публичного ключа
func ValidatePublicKey(publicKey string) error {
	if publicKey == "" {
		return errors.New("публичный ключ не может быть пустым")
	}

	// Проверяем, что это base64 строка
	_, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return fmt.Errorf("неверный формат публичного ключа: %w", err)
	}

	return nil
}

// ValidatePrivateKey проверяет корректность приватного ключа
func ValidatePrivateKey(privateKey string) error {
	if privateKey == "" {
		return errors.New("приватный ключ не может быть пустым")
	}

	// Проверяем, что это base64 строка
	_, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return fmt.Errorf("неверный формат приватного ключа: %w", err)
	}

	return nil
}

// SanitizeKeyPair очищает приватный ключ из структуры для передачи клиенту
func (kp *KeyPair) SanitizeForClient() map[string]interface{} {
	return map[string]interface{}{
		"public_key": kp.PublicKey,
		// Приватный ключ НЕ передается клиенту
	}
}

// IsValidKeyPair проверяет, что пара ключей корректна
func (kp *KeyPair) IsValid() error {
	if err := ValidatePublicKey(kp.PublicKey); err != nil {
		return fmt.Errorf("неверный публичный ключ: %w", err)
	}

	if err := ValidatePrivateKey(kp.PrivateKey); err != nil {
		return fmt.Errorf("неверный приватный ключ: %w", err)
	}

	return nil
}
