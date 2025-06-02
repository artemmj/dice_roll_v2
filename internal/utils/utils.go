package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// generateRandomHex генерирует криптографически безопасную
// случайную строку заданной длины в байтах
// Возвращает строку в HEX-формате (длина строки = length * 2)
func generateRandomHex(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// MustGenerateRandomHex - аналогична generateRandomHex, но паникует при ошибке
// Используется только в инициализациях где ошибки невозможны
func MustGenerateRandomHex(length int) string {
	s, err := generateRandomHex(length)
	if err != nil {
		panic(err)
	}
	return s
}
