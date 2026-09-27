package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	apphttp "github.com/AndewGitWorks/tamagochi/backend/internal/app/http"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
)

func main() {
	if err := loadProjectEnv(); err != nil {
		log.Fatal(err)
	}
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

// loadProjectEnv reads the root .env when the API is started from either the
// repository root or backend/. Existing shell environment values take priority.
func loadProjectEnv() error {
	err := loadEnvFile(".env")
	if errors.Is(err, os.ErrNotExist) {
		err = loadEnvFile("../.env")
	}
	return err
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found || strings.TrimSpace(key) == "" {
			return fmt.Errorf("invalid %s line %d", path, lineNumber)
		}
		if _, alreadySet := os.LookupEnv(key); !alreadySet {
			if err := os.Setenv(key, strings.TrimSpace(value)); err != nil {
				return fmt.Errorf("set %s from %s: %w", key, path, err)
			}
		}
	}
	return scanner.Err()
}
