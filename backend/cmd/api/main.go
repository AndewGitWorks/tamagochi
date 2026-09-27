package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	apphttp "github.com/AndewGitWorks/tamagochi/backend/internal/app/http"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET must be set and contain at least 32 characters")
	}

	service := users.NewAuthService(users.AuthConfig{
		Store:           users.NewMemoryStore(),
		JWTSecret:       []byte(secret),
		Issuer:          envOrDefault("JWT_ISSUER", "tamagochi-api"),
		AccessTokenTTL:  envDurationSeconds("JWT_ACCESS_TTL_SECONDS", 15*time.Minute),
		RefreshTokenTTL: envDurationSeconds("JWT_REFRESH_TTL_SECONDS", 30*24*time.Hour),
	})

	address := ":" + envOrDefault("PORT", "8080")
	log.Printf("Tamagochi API listening on %s", address)
	if err := http.ListenAndServe(address, apphttp.NewRouter(service)); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func envDurationSeconds(name string, fallback time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		log.Fatalf("%s must be a positive number of seconds", name)
	}
	return time.Duration(seconds) * time.Second
}
