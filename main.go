package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/nacl/box"
)

type Request struct {
	Password string `json:"password"`
}

type Response struct {
	Encrypted   string `json:"encrypted,omitempty"`
	OriginalEnc string `json:"originalenc,omitempty"`
	Error       string `json:"error,omitempty"`
}

func encryptHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Получаем пароль из query-параметра
	password := r.URL.Query().Get("password")
	if password == "" {
		respondError(w, "password parameter is required", http.StatusBadRequest)
		return
	}

	// Здесь должна быть логика получения publicKey и keyId с Facebook
	// Для примера используем фиксированные значения
	publicKey := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5fc3d46"
	keyId := 5

	result, err := encryptPassword(password, publicKey, keyId)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(
		Response{
			Encrypted:   result,
			OriginalEnc: "#PWD_BROWSER:10:1742826070:AZ9QAHhoSjryxAlcT96ezuvGCWuaaI3kCB/GXShBJPteRMEXAZoBZI9rXhGkO1hMx5Dn7vAKjv0uuGwrbkglUYOOJmXaZVrJRieOYgc5aaHKmRgTdkwuU1aPv5/n2vBdjIs9HaaMHLtnRvQlxQ==",
		})
}

func encryptPassword(password, publicKeyHex string, version int) (string, error) {
	// 1. Подготовка данных
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	plaintext := []byte(password)
	additionalData := []byte(timestamp)

	// 2. Парсинг публичного ключа
	publicKey, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid public key: %v", err)
	}
	if len(publicKey) != 32 {
		return "", fmt.Errorf("public key must be 32 bytes")
	}

	// 3. Генерация AES ключа
	aesKey := make([]byte, 32) // AES-256
	if _, err := rand.Read(aesKey); err != nil {
		return "", fmt.Errorf("failed to generate AES key: %v", err)
	}

	// 4. Шифрование AES ключа с помощью NaCl
	var recipientPubKey [32]byte
	copy(recipientPubKey[:], publicKey)

	sealedKey, err := box.SealAnonymous(nil, aesKey, &recipientPubKey, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to seal key: %v", err)
	}

	// 5. Шифрование пароля с AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %v", err)
	}

	encrypted := gcm.Seal(nil, iv, plaintext, additionalData)

	// 6. Формирование итогового буфера
	buffer := make([]byte, 100+len(plaintext))
	offset := 0

	buffer[offset] = 1
	offset++
	buffer[offset] = byte(version)
	offset++

	buffer[offset] = byte(len(sealedKey) & 255)
	buffer[offset+1] = byte(len(sealedKey) >> 8 & 255)
	offset += 2

	copy(buffer[offset:], sealedKey)
	offset += len(sealedKey)

	tag := encrypted[len(encrypted)-16:]
	ciphertext := encrypted[:len(encrypted)-16]

	copy(buffer[offset:], tag)
	offset += 16
	copy(buffer[offset:], ciphertext)

	// 7. Формирование результата
	return fmt.Sprintf("#PWD_BROWSER:10:%s:%s", timestamp, base64.StdEncoding.EncodeToString(buffer)), nil
}

func respondError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{Error: message})
}

func main() {
	http.HandleFunc("/encrypt", encryptHandler)

	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
