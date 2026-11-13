package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/yourusername/lambra/internal/models"
	"github.com/yourusername/lambra/internal/repository"
)

// AuthClaims holds JWT claims for an authenticated platform user
type AuthClaims struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtSecret  []byte
	tokenTTL   time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	if jwtSecret == "" {
		// Dev default; production must set JWT_SECRET
		jwtSecret = "lambra-dev-secret-change-me"
	}
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  24 * time.Hour,
	}
}

// Login verifies credentials and returns a signed JWT with the user
func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to issue token: %w", err)
	}

	// Never expose the hash, even in memory
	user.PasswordHash = ""
	return &models.LoginResponse{Token: token, User: user}, nil
}

// generateToken signs a JWT for the given user
func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := AuthClaims{
		UUID:     user.UUID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.UUID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "lambra",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken parses and validates a JWT, returning its claims
func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
