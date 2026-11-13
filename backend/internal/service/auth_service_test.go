package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

// connectTestDB connects to the local Lambra database for DB-backed tests
func connectTestDB() (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		getEnvOrDefault("DB_USER", "lambra"),
		getEnvOrDefault("DB_PASSWORD", "lambra_secret"),
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "3306"),
		getEnvOrDefault("DB_NAME", "lambra_db"),
	)
	return sqlx.Connect("mysql", dsn)
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func setupTestAuthService() (*AuthService, *repository.UserRepository, error) {
	db, err := connectTestDB()
	if err != nil {
		return nil, nil, err
	}
	userRepo := repository.NewUserRepository(db)
	return NewAuthService(userRepo, "test-secret"), userRepo, nil
}

func TestAuthService_Login(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DB-dependent test in short mode")
	}

	svc, _, err := setupTestAuthService()
	if err != nil {
		t.Skipf("test DB unavailable: %v", err)
	}

	req := &models.LoginRequest{Username: "admin", Password: "admin123"}
	result, err := svc.Login(req)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.Token == "" {
		t.Error("Login() returned empty token")
	}
	if result.User == nil || result.User.Username != "admin" {
		t.Errorf("Login() user = %+v, want admin", result.User)
	}
	if result.User.PasswordHash != "" {
		t.Error("Login() exposed password hash in response")
	}
	userJSON, _ := json.Marshal(result.User)
	if strings.Contains(string(userJSON), "password_hash") {
		t.Errorf("Login() user JSON exposes password hash: %s", userJSON)
	}
}

func TestAuthService_LoginWrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DB-dependent test in short mode")
	}

	svc, _, err := setupTestAuthService()
	if err != nil {
		t.Skipf("test DB unavailable: %v", err)
	}

	req := &models.LoginRequest{Username: "admin", Password: "wrong"}
	if _, err := svc.Login(req); err == nil {
		t.Error("Login() with wrong password should fail")
	}
}

func TestAuthService_TokenRoundTrip(t *testing.T) {
	svc := NewAuthService(nil, "round-trip-secret")

	user := &models.User{UUID: "abc-123", Username: "admin", Role: "admin"}
	token, err := svc.generateToken(user)
	if err != nil {
		t.Fatalf("generateToken() error = %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims["username"] != "admin" || claims["uuid"] != "abc-123" || claims["role"] != "admin" {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestAuthService_ValidateTokenRejectsWrongSecret(t *testing.T) {
	signer := NewAuthService(nil, "secret-one")
	verifier := NewAuthService(nil, "secret-two")

	token, err := signer.generateToken(&models.User{UUID: "u", Username: "admin", Role: "admin"})
	if err != nil {
		t.Fatalf("generateToken() error = %v", err)
	}

	if _, err := verifier.ValidateToken(token); err == nil {
		t.Error("ValidateToken() should reject a token signed with a different secret")
	}
}
