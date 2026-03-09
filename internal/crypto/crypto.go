// Package crypto предоставляет функции для асимметричного шифрования с использованием RSA.
//
// Пакет реализует:
//   - Загрузку RSA ключей из файлов и io.Reader
//   - Поддержку публичных и приватных ключей в форматах PEM
//   - Совместимость с ключами в форматах PKCS#1 и PKCS#8
//
// Основные функции:
//   - LoadRSAPublicKey / ReadRSAPublicKey: загрузка публичных ключей
//   - LoadRSAPrivateKey / ReadRSAPrivateKey: загрузка приватных ключей
//   - EncryptWithPublicKey: шифрование данных публичным ключом
//   - DecryptWithPrivateKey: расшифрование данных приватным ключом
//
// Пример использования:
//
//	// Загрузка ключей
//	publicKey, err := crypto.LoadRSAPublicKey("public.pem")
//	privateKey, err := crypto.LoadRSAPrivateKey("private.pem")
//
//	// Шифрование и расшифрование
//	ciphertext, err := crypto.EncryptWithPublicKey(publicKey, []byte("secret"))
//	plaintext, err := crypto.DecryptWithPrivateKey(privateKey, ciphertext)
package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

// LoadRSAPublicKey загружает RSA публичный ключ из файла по указанному пути.
//
// Функция открывает файл, читает его содержимое и передает его в ReadRSAPublicKey
// для парсинга. Поддерживает ключи в формате PEM.
//
// Параметры:
//   - filename: путь к файлу, содержащему публичный ключ в формате PEM
//
// Возвращает:
//   - *rsa.PublicKey: указатель на загруженный публичный ключ
//   - error: ошибка, если файл не найден, поврежден или не является валидным RSA ключом
func LoadRSAPublicKey(filename string) (*rsa.PublicKey, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	return ReadRSAPublicKey(file)
}

// ReadRSAPublicKey читает и парсит RSA публичный ключ из любого источника, реализующего io.Reader.
//
// Функция читает все данные из reader, декодирует PEM-блок и извлекает публичный ключ
// в формате PKIX (X.509). Поддерживает стандартные PEM-обертки типа "RSA PUBLIC KEY".
//
// Параметры:
//   - r: io.Reader, из которого будут прочитаны данные ключа (например, *os.File, bytes.Reader)
//
// Возвращает:
//   - *rsa.PublicKey: указатель на распарсенный публичный ключ
//   - error: ошибка, если данные повреждены, не содержат PEM-блока или не являются валидным RSA ключом
func ReadRSAPublicKey(r io.Reader) (*rsa.PublicKey, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения данных: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("не удалось декодировать PEM блок")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга публичного ключа: %w", err)
	}

	publicKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("ключ не является RSA публичным ключом")
	}

	return publicKey, nil
}

// LoadRSAPrivateKey загружает RSA приватный ключ из файла по указанному пути.
//
// Функция открывает файл и передает его содержимое в ReadRSAPrivateKey для парсинга.
// Поддерживает приватные ключи в форматах PKCS#1 и PKCS#8 (PEM).
//
// Параметры:
//   - filename: путь к файлу, содержащему приватный ключ в формате PEM
//
// Возвращает:
//   - *rsa.PrivateKey: указатель на загруженный приватный ключ
//   - error: ошибка, если файл не найден, поврежден, защищен паролем или не является валидным RSA ключом
func LoadRSAPrivateKey(filename string) (*rsa.PrivateKey, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	return ReadRSAPrivateKey(file)
}

// ReadRSAPrivateKey читает и парсит RSA приватный ключ из любого источника, реализующего io.Reader.
//
// Функция поддерживает два формата хранения приватных ключей:
//   - PKCS#1 (BEGIN RSA PRIVATE KEY)
//   - PKCS#8 (BEGIN PRIVATE KEY)
//
// Сначала пытается распарсить ключ как PKCS#1, при неудаче пробует PKCS#8.
// Это обеспечивает совместимость с большинством генераторов ключей.
//
// Параметры:
//   - r: io.Reader, из которого будут прочитаны данные ключа
//
// Возвращает:
//   - *rsa.PrivateKey: указатель на распарсенный приватный ключ
//   - error: ошибка, если данные повреждены, не содержат PEM-блока, защищены паролем или не являются валидным RSA ключом
func ReadRSAPrivateKey(r io.Reader) (*rsa.PrivateKey, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения данных: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("не удалось декодировать PEM блок")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Возможно, ключ в формате PKCS8
		key8, err8 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err8 != nil {
			return nil, fmt.Errorf("ошибка парсинга приватного ключа (PKCS1 и PKCS8): %w", err)
		}
		if privateKey, ok := key8.(*rsa.PrivateKey); ok {
			return privateKey, nil
		}
		return nil, fmt.Errorf("ключ не является RSA приватным ключом")
	}

	return key, nil
}
