package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	saltLabel = "postgres-connector-salt"
	keyLen    = 32
	ivLen     = 16
	saltLen   = 64
	tagLen    = 16
	aadValue  = "postgres-connector"
	sN        = 16384
	sR        = 8
	sP        = 1
)

func deriveKey(master string) ([]byte, error) {
	return scrypt.Key([]byte(master), []byte(saltLabel), sN, sR, sP, keyLen)
}

// Encrypt matches apps/api encryption.service.ts (AES-256-GCM, format salt:iv:tag:ciphertext base64 parts).
func Encrypt(masterKey, plaintext string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("plaintext required")
	}
	key, err := deriveKey(masterKey)
	if err != nil {
		return "", err
	}
	salt := make([]byte, saltLen)
	iv := make([]byte, ivLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, ivLen)
	if err != nil {
		return "", err
	}
	ct := gcm.Seal(nil, iv, []byte(plaintext), []byte(aadValue))
	if len(ct) < tagLen {
		return "", fmt.Errorf("ciphertext too short")
	}
	ciphertext := ct[:len(ct)-tagLen]
	tag := ct[len(ct)-tagLen:]
	return fmt.Sprintf("%s:%s:%s:%s",
		base64.StdEncoding.EncodeToString(salt),
		base64.StdEncoding.EncodeToString(iv),
		base64.StdEncoding.EncodeToString(tag),
		base64.StdEncoding.EncodeToString(ciphertext),
	), nil
}

// Decrypt reverses Encrypt / Nest encryption.service decrypt.
func Decrypt(masterKey, encrypted string) (string, error) {
	parts := strings.Split(encrypted, ":")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid encryption format")
	}
	salt, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil || len(salt) != saltLen {
		return "", fmt.Errorf("invalid salt")
	}
	_ = salt
	iv, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil || len(iv) != ivLen {
		return "", fmt.Errorf("invalid iv")
	}
	tag, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil || len(tag) != tagLen {
		return "", fmt.Errorf("invalid tag")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext")
	}
	key, err := deriveKey(masterKey)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, ivLen)
	if err != nil {
		return "", err
	}
	data := append(ciphertext, tag...)
	pt, err := gcm.Open(nil, iv, data, []byte(aadValue))
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(pt), nil
}
