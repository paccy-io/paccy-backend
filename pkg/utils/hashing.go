package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, error) {
	// Her kullanıcı için benzersiz salt üret
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Argon2id ile hashle
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Salt + Hash'i birlikte sakla (Base64 ile)
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	// Format: salt$hash
	return fmt.Sprintf("%s$%s", saltB64, hashB64), nil
}

func VerifyPassword(hashedPassword, plainPassword string) bool {
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 2 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	hash := argon2.IDKey([]byte(plainPassword), salt, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(hash, expectedHash) == 1
}
