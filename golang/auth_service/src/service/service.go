package service

import (
	"errors"
	"fmt"
	"time"

	"auth_service/src/repository"
	shared "shared/auth"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
	tokenTTL  time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		tokenTTL:  tokenTTL,
	}
}

// Register создаёт нового пользователя и возвращает JWT
func (s *AuthService) Register(username, email, password string) (shared.AuthResponse, error) {
	userID, err := s.userRepo.Create(username, email, password)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEmailAlreadyExists):
			return shared.AuthResponse{Success: false, Error: "Email already registered"}, err
		case errors.Is(err, repository.ErrUsernameAlreadyExists):
			return shared.AuthResponse{Success: false, Error: "Username already taken"}, err
		default:
			return shared.AuthResponse{Success: false, Error: "Failed to create user"}, err
		}
	}

	token, expiresAt, err := s.generateToken(userID, email)
	if err != nil {
		return shared.AuthResponse{Success: false, Error: "Failed to generate token"}, err
	}

	return shared.AuthResponse{
		Success:   true,
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		UserID:    userID,
	}, nil
}

// Login аутентифицирует пользователя по email и паролю
// Аналог get_user_uuid из Rust-кода
func (s *AuthService) Login(email, password string) (shared.AuthResponse, error) {
	// 1. Получаем пользователя по email (аналог get_user_by_email)
	userID, passwordHash, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return shared.AuthResponse{Success: false, Error: "Invalid credentials"}, err
		}
		return shared.AuthResponse{Success: false, Error: "Authentication failed"}, err
	}

	// 2. Проверяем пароль (аналог compare_passwords)
	passwordValid, err := s.userRepo.ComparePasswords(password, passwordHash)
	if err != nil {
		return shared.AuthResponse{Success: false, Error: "Authentication failed"}, err
	}

	if !passwordValid {
		return shared.AuthResponse{Success: false, Error: "Invalid credentials"}, repository.ErrInvalidCredentials
	}

	// 3. Генерируем JWT (аналог get_jwt_token)
	token, expiresAt, err := s.generateToken(userID.String(), email)
	if err != nil {
		return shared.AuthResponse{Success: false, Error: "Failed to generate token"}, err
	}

	return shared.AuthResponse{
		Success:   true,
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		UserID:    userID.String(),
	}, nil
}

// generateToken создаёт JWT с UUID пользователя и сроком действия
func (s *AuthService) generateToken(userID, email string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.tokenTTL)

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}
