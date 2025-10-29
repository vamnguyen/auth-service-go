package service

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"auth-service/internal/model"
	"auth-service/internal/repository"
	"auth-service/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	UserRepo    *repository.UserRepository
	RefreshRepo *repository.RefreshTokenRepository
	JWTSecret   string
}

func NewAuthService(userRepo *repository.UserRepository, refreshRepo *repository.RefreshTokenRepository, jwtSecret string) *AuthService {
	return &AuthService{
		UserRepo:    userRepo,
		RefreshRepo: refreshRepo,
		JWTSecret:   jwtSecret,
	}
}

func (s *AuthService) Register(email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	return s.UserRepo.CreateUser(user)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.UserRepo.FindUserByEmail(email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID.String(), s.JWTSecret, 24*time.Hour)
	return token, err
}

// GetMe returns the current user profile by userID (string UUID)
func (s *AuthService) GetMe(userID string) (*model.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	return s.UserRepo.FindUserByID(id)
}

// Logout revokes all refresh tokens of the current user (stateless access token remains valid until expiry)
func (s *AuthService) Logout(userID string) error {
	if s.RefreshRepo == nil {
		return fmt.Errorf("refresh token repository not initialized")
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	return s.RefreshRepo.RevokeAllByUser(id)
}
