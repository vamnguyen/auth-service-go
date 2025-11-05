package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret         string
	accessTokenTTL time.Duration
}

func NewJWTService(secret string, accessTokenTTL time.Duration) *JWTService {
	return &JWTService{
		secret:         secret,
		accessTokenTTL: accessTokenTTL,
	}
}

func (s *JWTService) GenerateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.accessTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *JWTService) GenerateRefreshToken() (plain, hash string, err error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}

	plain = base64.RawURLEncoding.EncodeToString(randomBytes)
	hash = s.HashToken(plain)
	return plain, hash, nil
}

func (s *JWTService) HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (s *JWTService) ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	if expVal, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(expVal), 0).Before(time.Now()) {
			return "", errors.New("token expired")
		}
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", errors.New("user_id missing in token")
	}

	return userID, nil
}
