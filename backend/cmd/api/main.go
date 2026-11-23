package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AndewGitWorks/tamagochi/backend/internal/app/http/routes"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/orders"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/products"
	"github.com/AndewGitWorks/tamagochi/backend/internal/domain/users"
	"github.com/AndewGitWorks/tamagochi/backend/internal/platform/database"
	"github.com/AndewGitWorks/tamagochi/backend/internal/repository"
)

func main() {
	if err := loadProjectEnv(); err != nil {
		log.Fatal(err)
	}
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET must be set and contain at least 32 characters")
	}
	dsn, err := databaseURLFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.OpenPostgres(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	service := users.NewAuthService(users.AuthConfig{Store: repository.NewPostgresUserStore(db), RefreshSessions: repository.NewPostgresRefreshSessionStore(db), JWTSecret: []byte(secret),
		Issuer: envOrDefault("JWT_ISSUER", "tamagochi-api"), AccessTokenTTL: envDurationSeconds("JWT_ACCESS_TTL_SECONDS", 15*time.Minute),
		RefreshTokenTTL: envDurationSeconds("JWT_REFRESH_TTL_SECONDS", 30*24*time.Hour)})
	address := ":" + envOrDefault("PORT", "8080")
	log.Printf("Tamagochi API listening on %s", address)
	userStore := repository.NewPostgresUserStore(db)
	productStore := repository.NewPostgresProductStore(db)
	imageStore := repository.NewPostgresProductImageStore(db)
	if err := http.ListenAndServe(address, routes.NewRouter(service, routes.Config{
		Users:          userStore,
		UserService:    users.NewUserService(userStore),
		ProfileService: users.NewProfileService(repository.NewPostgresProfileStore(db)),
		CatalogService: products.NewService(productStore, imageStore),
		OrderService: orders.NewService(
			repository.NewPostgresOrderStore(db),
			repository.NewPostgresOrderItemStore(db),
		),
		FavoriteService: products.NewFavoriteService(productStore, repository.NewPostgresLikedProductStore(db)),
	})); err != nil {
		log.Fatal(err)
	}
}

func databaseURLFromEnv() (string, error) {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn, nil
	}
	user, password, dbName := os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB")
	if user == "" || password == "" || dbName == "" {
		return "", fmt.Errorf("set DATABASE_URL or POSTGRES_USER, POSTGRES_PASSWORD, and POSTGRES_DB")
	}
	host := envOrDefault("POSTGRES_HOST", "localhost")
	port := envOrDefault("POSTGRES_PORT", "5432")
	sslMode := envOrDefault("POSTGRES_SSLMODE", "disable")
	u := url.URL{Scheme: "postgres", User: url.UserPassword(user, password), Host: host + ":" + port, Path: "/" + dbName}
	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()
	return u.String(), nil
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
