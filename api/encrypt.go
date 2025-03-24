package handler

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"golang.org/x/crypto/nacl/box"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	password := r.URL.Query().Get("password")
	if password == "" {
		http.Error(w, "password parameter is required", http.StatusBadRequest)
		return
	}

	// Фиксированные значения для примера
	publicKeyHex := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6"
	version := 10

	result, err := encrypt(password, publicKeyHex, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(result))
}

func encrypt(password, publicKeyHex string, version int) (string, error) {
	publicKey, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return "", err
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	plaintext := []byte(password)
	additionalData := []byte(timestamp)

	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return "", err
	}

	var recipientPubKey [32]byte
	copy(recipientPubKey[:], publicKey)

	sealedKey, err := box.SealAnonymous(nil, aesKey, &recipientPubKey, rand.Reader)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	encrypted := gcm.Seal(nil, iv, plaintext, additionalData)

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

	return "#PWD_BROWSER:" + strconv.Itoa(version) + ":" + timestamp + ":" + base64.StdEncoding.EncodeToString(buffer), nil
}
