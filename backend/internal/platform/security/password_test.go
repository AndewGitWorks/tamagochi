package security

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestHashPasswordFitsDatabaseColumnAndVerifies(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) > 64 {
		t.Fatalf("encoded hash length %d exceeds password_hashed VARCHAR(64)", len(hash))
	}
	if !VerifyPassword("correct-horse-battery-staple", hash) {
		t.Fatal("password hash did not verify")
	}
	if VerifyPassword("wrong-password", hash) {
		t.Fatal("wrong password verified")
	}
}

func TestHashPasswordUsesDifferentSalt(t *testing.T) {
	first, err := HashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("password hashes should use independently generated salts")
	}
}

func TestVerifyPasswordSupportsLegacyEncoding(t *testing.T) {
	salt := []byte("0123456789abcdef")
	key := pbkdf2SHA256([]byte("legacy-password"), salt, 1, 32)
	legacy := fmt.Sprintf("pbkdf2-sha256$1$%s$%s", base64.RawURLEncoding.EncodeToString(salt), base64.RawURLEncoding.EncodeToString(key))
	if !VerifyPassword("legacy-password", legacy) {
		t.Fatal("legacy hash did not verify")
	}
	if VerifyPassword("wrong-password", legacy) {
		t.Fatal("wrong password matched legacy hash")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	if VerifyPassword("anything", "not-a-password-hash") {
		t.Fatal("malformed hash verified")
	}
}
