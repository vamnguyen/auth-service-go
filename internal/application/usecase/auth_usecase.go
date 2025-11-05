package usecase

import (
	"context"
	"time"

	"auth-service/internal/application/dto"
	"auth-service/internal/domain/entity"
	domainErr "auth-service/internal/domain/error"
	"auth-service/internal/domain/repository"

	"github.com/google/uuid"
)

type AuthUseCase struct {
	userRepo        repository.UserRepository
	refreshRepo     repository.RefreshTokenRepository
	auditRepo       repository.AuditLogRepository
	tokenService    TokenService
	passwordService PasswordService
	config          AuthConfig
}

type AuthConfig struct {
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	MaxLoginAttempts    int
	AccountLockDuration time.Duration
}

type TokenService interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken() (plain, hash string, err error)
	HashToken(plain string) string
	ValidateAccessToken(token string) (userID string, err error)
}

type PasswordService interface {
	ValidateStrength(password string) error
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	auditRepo repository.AuditLogRepository,
	tokenService TokenService,
	passwordService PasswordService,
	config AuthConfig,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:        userRepo,
		refreshRepo:     refreshRepo,
		auditRepo:       auditRepo,
		tokenService:    tokenService,
		passwordService: passwordService,
		config:          config,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, req dto.RegisterRequest) error {
	exists, err := uc.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if exists {
		return domainErr.ErrUserAlreadyExists
	}

	if err := uc.passwordService.ValidateStrength(req.Password); err != nil {
		return err
	}

	user, err := entity.NewUser(req.Email, req.Password)
	if err != nil {
		return err
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return err
	}

	auditLog := entity.NewAuditLog(user.ID, entity.AuditActionRegister, "", "")
	_ = uc.auditRepo.Create(ctx, auditLog)

	return nil
}

func (uc *AuthUseCase) Login(ctx context.Context, req dto.LoginRequest, ipAddress, userAgent string) (*dto.LoginResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		auditLog := entity.NewAuditLog(entity.User{}.ID, entity.AuditActionLoginFailed, ipAddress, userAgent)
		auditLog.AddMetadata("email", req.Email)
		_ = uc.auditRepo.Create(ctx, auditLog)
		return nil, domainErr.ErrInvalidCredentials
	}

	if user.IsAccountLocked() {
		auditLog := entity.NewAuditLog(user.ID, entity.AuditActionAccountLocked, ipAddress, userAgent)
		_ = uc.auditRepo.Create(ctx, auditLog)
		return nil, domainErr.ErrAccountLocked
	}

	if err := user.VerifyPassword(req.Password); err != nil {
		user.IncrementFailedLoginAttempts(uc.config.MaxLoginAttempts, uc.config.AccountLockDuration)
		_ = uc.userRepo.Update(ctx, user)

		auditLog := entity.NewAuditLog(user.ID, entity.AuditActionLoginFailed, ipAddress, userAgent)
		_ = uc.auditRepo.Create(ctx, auditLog)

		return nil, domainErr.ErrInvalidCredentials
	}

	user.ResetFailedLoginAttempts()
	user.UpdateLastLogin(ipAddress)
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	refreshPlain, refreshHash, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshToken := entity.NewRefreshToken(user.ID, refreshHash, uc.config.RefreshTokenTTL)
	if err := uc.refreshRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	auditLog := entity.NewAuditLog(user.ID, entity.AuditActionLogin, ipAddress, userAgent)
	_ = uc.auditRepo.Create(ctx, auditLog)

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshPlain,
		TokenType:    "Bearer",
		ExpiresIn:    int(uc.config.AccessTokenTTL.Seconds()),
		User: dto.UserDTO{
			ID:         user.ID.String(),
			Email:      user.Email,
			Role:       string(user.Role),
			IsVerified: user.IsVerified,
			CreatedAt:  user.CreatedAt,
		},
	}, nil
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshPlain string) (*dto.RefreshTokenResponse, error) {
	if refreshPlain == "" {
		return nil, domainErr.ErrMissingToken
	}

	refreshHash := uc.tokenService.HashToken(refreshPlain)

	token, err := uc.refreshRepo.FindByTokenHash(ctx, refreshHash)
	if err != nil {
		return nil, domainErr.ErrInvalidToken
	}

	if !token.IsValid() {
		return nil, domainErr.ErrTokenExpired
	}

	if err := uc.refreshRepo.RevokeByTokenHash(ctx, refreshHash); err != nil {
		return nil, err
	}

	newAccessToken, err := uc.tokenService.GenerateAccessToken(token.UserID.String())
	if err != nil {
		return nil, err
	}

	newRefreshPlain, newRefreshHash, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	newRefreshToken := entity.NewRefreshToken(token.UserID, newRefreshHash, uc.config.RefreshTokenTTL)
	if err := uc.refreshRepo.Create(ctx, newRefreshToken); err != nil {
		return nil, err
	}

	auditLog := entity.NewAuditLog(token.UserID, entity.AuditActionTokenRefresh, "", "")
	_ = uc.auditRepo.Create(ctx, auditLog)

	return &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshPlain,
		TokenType:    "Bearer",
		ExpiresIn:    int(uc.config.AccessTokenTTL.Seconds()),
	}, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, userID string, refreshPlain string, ipAddress, userAgent string) error {
	if refreshPlain != "" {
		refreshHash := uc.tokenService.HashToken(refreshPlain)
		_ = uc.refreshRepo.RevokeByTokenHash(ctx, refreshHash)
	}

	if userID != "" {
		userUUID, err := uuid.Parse(userID)
		if err == nil {
			auditLog := entity.NewAuditLog(userUUID, entity.AuditActionLogout, ipAddress, userAgent)
			_ = uc.auditRepo.Create(ctx, auditLog)
		}
	}

	return nil
}

func (uc *AuthUseCase) LogoutAll(ctx context.Context, userID string, ipAddress, userAgent string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return domainErr.ErrInvalidInput
	}

	if err := uc.refreshRepo.RevokeAllByUserID(ctx, userUUID); err != nil {
		return err
	}

	auditLog := entity.NewAuditLog(userUUID, entity.AuditActionLogout, ipAddress, userAgent)
	auditLog.AddMetadata("all_sessions", true)
	_ = uc.auditRepo.Create(ctx, auditLog)

	return nil
}

func (uc *AuthUseCase) GetMe(ctx context.Context, userID string) (*dto.UserDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domainErr.ErrInvalidInput
	}

	user, err := uc.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		return nil, domainErr.ErrUserNotFound
	}

	return &dto.UserDTO{
		ID:         user.ID.String(),
		Email:      user.Email,
		Role:       string(user.Role),
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
	}, nil
}

func (uc *AuthUseCase) ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return domainErr.ErrInvalidInput
	}

	user, err := uc.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		return domainErr.ErrUserNotFound
	}

	if err := user.VerifyPassword(req.OldPassword); err != nil {
		return domainErr.ErrInvalidPassword
	}

	if err := uc.passwordService.ValidateStrength(req.NewPassword); err != nil {
		return err
	}

	if err := user.ChangePassword(req.NewPassword); err != nil {
		return err
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return err
	}

	_ = uc.refreshRepo.RevokeAllByUserID(ctx, userUUID)

	auditLog := entity.NewAuditLog(user.ID, entity.AuditActionPasswordChange, "", "")
	_ = uc.auditRepo.Create(ctx, auditLog)

	return nil
}
