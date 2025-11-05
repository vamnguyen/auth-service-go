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
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
}

func NewAuthService(
	userRepo *repository.UserRepository,
	refreshRepo *repository.RefreshTokenRepository,
	jwtSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		UserRepo:    userRepo,
		RefreshRepo: refreshRepo,
		JWTSecret:   jwtSecret,
		AccessTTL:   accessTTL,
		RefreshTTL:  refreshTTL,
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

func (s *AuthService) Login(email, password string) (string, string, error) {
	user, err := s.UserRepo.FindUserByEmail(email)
	if err != nil {
		return "", "", errors.New("user not found")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", "", errors.New("invalid credentials")
	}

	// Access token (short)
	accessToken, err := utils.GenerateToken(user.ID.String(), s.JWTSecret, s.AccessTTL)
	if err != nil {
		return "", "", err
	}

	// Refresh token (long, store HASH)
	refreshPlain, refreshHash, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	rt := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshHash, // store HASH, NOT plain
		ExpiresAt: time.Now().Add(s.RefreshTTL),
		Revoked:   false,
	}
	if err := s.RefreshRepo.Create(rt); err != nil {
		return "", "", err
	}

	return accessToken, refreshPlain, nil
}

func (s *AuthService) Refresh(refreshPlain string, userIDExpected string) (string, string, error) {
	if refreshPlain == "" {
		return "", "", errors.New("missing refresh token")
	}
	refreshHash := utils.HashRefreshToken(refreshPlain)

	rt, err := s.RefreshRepo.FindByTokenHash(refreshHash)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}
	if rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return "", "", errors.New("refresh token expired or revoked")
	}
	// Optional: enforce refresh belongs to the same user (defense-in-depth)
	if userIDExpected != "" && rt.UserID.String() != userIDExpected {
		return "", "", errors.New("refresh token does not belong to user")
	}

	// Rotate: revoke old
	_ = s.RefreshRepo.RevokeByTokenHash(refreshHash)

	// Issue new pair
	newAccess, err := utils.GenerateToken(rt.UserID.String(), s.JWTSecret, s.AccessTTL)
	if err != nil {
		return "", "", err
	}
	newRefreshPlain, newRefreshHash, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	newRT := &model.RefreshToken{
		UserID:    rt.UserID,
		Token:     newRefreshHash,
		ExpiresAt: time.Now().Add(s.RefreshTTL),
		Revoked:   false,
	}
	if err := s.RefreshRepo.Create(newRT); err != nil {
		return "", "", err
	}

	return newAccess, newRefreshPlain, nil
}

func (s *AuthService) LogoutCurrent(refreshPlain string) error {
	if refreshPlain == "" {
		return errors.New("missing refresh token")
	}
	refreshHash := utils.HashRefreshToken(refreshPlain)
	return s.RefreshRepo.RevokeByTokenHash(refreshHash)
}

// Logout revokes ALL refresh tokens of the user
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

func (s *AuthService) GetMe(userID string) (*model.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	return s.UserRepo.FindUserByID(id)
}
