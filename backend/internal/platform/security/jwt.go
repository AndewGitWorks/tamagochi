package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("bearer token is invalid or has expired")

type Claims struct {
	Issuer         string `json:"iss"`
	Subject        string `json:"sub"`
	Login          string `json:"login"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	TimeExpiration string `json:"time_expiration"`
	TimeCreation   string `json:"time_creation"`
	IssuedAt       int64  `json:"iat"`
	ExpiresAt      int64  `json:"exp"`
	ID             string `json:"jti"`
	TokenType      string `json:"token_type"`
}

func SignJWT(claims Claims, secret []byte) (string, error) {
	header, err := json.Marshal(struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}{Algorithm: "HS256", Type: "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	return unsigned + "." + signature(unsigned, secret), nil
}

func VerifyJWT(token string, secret []byte, issuer, expectedType string, now time.Time) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(signature(parts[0]+"."+parts[1], secret))) != 1 {
		return Claims{}, ErrInvalidToken
	}

	var header struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || json.Unmarshal(headerJSON, &header) != nil || header.Algorithm != "HS256" || header.Type != "JWT" {
		return Claims{}, ErrInvalidToken
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var claims Claims
	if json.Unmarshal(payloadJSON, &claims) != nil ||
		claims.Issuer != issuer || claims.TokenType != expectedType ||
		claims.Subject == "" || claims.Login == "" || claims.Name == "" || claims.Role == "" || claims.ID == "" ||
		claims.IssuedAt == 0 || claims.ExpiresAt <= now.Unix() {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}

func signature(unsigned string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
