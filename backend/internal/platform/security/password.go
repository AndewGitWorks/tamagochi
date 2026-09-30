package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const passwordIterations = 210_000
const passwordKeyLength = 24 // 192-bit key; keeps the complete encoding within VARCHAR(64).
const passwordSaltLength = 16
const passwordHashVersion = 1
const passwordIterationBytes = 3

// HashPassword uses PBKDF2-HMAC-SHA256 with a unique random salt. The compact
// versioned encoding stores version, iteration count, salt, and derived key in
// one base64 value so it fits the existing password_hashed VARCHAR(64) column.
func HashPassword(password string) (string, error) {
	salt := make([]byte, passwordSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("create password salt: %w", err)
	}
	derived := pbkdf2SHA256([]byte(password), salt, passwordIterations, passwordKeyLength)
	encoded := make([]byte, 1+passwordIterationBytes+len(salt)+len(derived))
	encoded[0] = passwordHashVersion
	encoded[1] = byte(passwordIterations >> 16)
	encoded[2] = byte(passwordIterations >> 8 & 0xff)
	encoded[3] = byte(passwordIterations & 0xff)
	copy(encoded[4:], salt)
	copy(encoded[4+len(salt):], derived)
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func VerifyPassword(password, encoded string) bool {
	// Keep support for hashes created by the previous textual representation.
	if strings.HasPrefix(encoded, "pbkdf2-sha256$") {
		return verifyLegacyPassword(password, encoded)
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(data) != 1+passwordIterationBytes+passwordSaltLength+passwordKeyLength || data[0] != passwordHashVersion {
		return false
	}
	iterations := int(data[1])<<16 | int(data[2])<<8 | int(data[3])
	if iterations < 1 {
		return false
	}
	salt := data[4 : 4+passwordSaltLength]
	expected := data[4+passwordSaltLength:]
	actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(expected, actual) == 1
}

func verifyLegacyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations < 1 {
		return false
	}
	salt, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	expected, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil || len(expected) == 0 {
		return false
	}
	actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(expected, actual) == 1
}

func pbkdf2SHA256(password, salt []byte, iterations, length int) []byte {
	result := make([]byte, 0, length)
	for block := uint32(1); len(result) < length; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for round := 1; round < iterations; round++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for i := range t {
				t[i] ^= u[i]
			}
		}
		result = append(result, t...)
	}
	return result[:length]
}
