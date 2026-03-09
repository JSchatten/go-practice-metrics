package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
)

// generateAESKey создает случайный 256-битный ключ AES.
func generateAESKey() ([]byte, error) {
	key := make([]byte, 32) // 256 бит = 32 байта
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать AES ключ: %w", err)
	}
	return key, nil
}

// encryptAES шифрует данные с использованием AES-GCM.
func encryptAES(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания AES шифра: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("ошибка генерации nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decryptAES расшифровывает данные, зашифрованные AES-GCM.
func decryptAES(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания AES шифра: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GCM: %w", err)
	}

	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("данные повреждены: недостаточная длина")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// HybridEncrypt выполняет гибридное шифрование: шифрует данные с помощью AES,
// а ключ AES шифрует с помощью RSA публичного ключа.
//
// Формат результата: base64(encryptedAESKey || encryptedData)
func HybridEncrypt(pub *rsa.PublicKey, data []byte) (string, error) {
	// Генерируем случайный AES ключ
	aesKey, err := generateAESKey()
	if err != nil {
		return "", err
	}

	// Шифруем данные с помощью AES
	encryptedData, err := encryptAES(aesKey, data)
	if err != nil {
		return "", err
	}

	// Шифруем AES ключ с помощью RSA
	encryptedAESKey, err := EncryptWithPublicKey(pub, aesKey)
	if err != nil {
		return "", fmt.Errorf("ошибка шифрования AES ключа: %w", err)
	}

	// Объединяем зашифрованный ключ и данные
	result := append(encryptedAESKey, encryptedData...)

	// Кодируем в base64 для передачи в HTTP
	return base64.StdEncoding.EncodeToString(result), nil
}

// HybridDecrypt выполняет гибридное расшифрование: расшифровывает AES ключ с помощью RSA,
// затем расшифровывает данные с помощью AES.
//
// Принимает данные в формате: base64(encryptedAESKey || encryptedData)
func HybridDecrypt(priv *rsa.PrivateKey, data string) ([]byte, error) {
	// Декодируем из base64
	combinedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования base64: %w", err)
	}

	// RSA ключи 2048 бит = 256 байт
	// Это хардкод, но для учебного проекта подойдет
	// В продакшене нужно знать длину ключа
	encryptedKeySize := priv.PublicKey.Size()

	if len(combinedData) < encryptedKeySize {
		return nil, fmt.Errorf("данные повреждены: недостаточная длина")
	}

	encryptedAESKey := combinedData[:encryptedKeySize]
	encryptedData := combinedData[encryptedKeySize:]

	// Расшифровываем AES ключ с помощью RSA
	aesKey, err := DecryptWithPrivateKey(priv, encryptedAESKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифрования AES ключа: %w", err)
	}

	// Расшифровываем данные с помощью AES
	return decryptAES(aesKey, encryptedData)
}
