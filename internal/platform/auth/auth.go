package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}
type TokenManager struct {
	accessSecret, refreshSecret []byte
	accessTTL, refreshTTL       time.Duration
}

func NewTokenManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{[]byte(accessSecret), []byte(refreshSecret), accessTTL, refreshTTL}
}
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
func (m *TokenManager) Issue(userID, email string) (string, string, error) {
	access, err := m.issue(userID, email, "access", m.accessTTL, m.accessSecret)
	if err != nil {
		return "", "", err
	}
	refresh, err := m.issue(userID, email, "refresh", m.refreshTTL, m.refreshSecret)
	return access, refresh, err
}
func (m *TokenManager) issue(userID, email, tokenType string, ttl time.Duration, secret []byte) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{UserID: userID, Email: email, TokenType: tokenType, RegisteredClaims: jwt.RegisteredClaims{Subject: userID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(ttl))}})
	return token.SignedString(secret)
}
func (m *TokenManager) ParseAccess(tokenString string) (*Claims, error) {
	return m.parse(tokenString, m.accessSecret, "access")
}
func (m *TokenManager) ParseRefresh(tokenString string) (*Claims, error) {
	return m.parse(tokenString, m.refreshSecret, "refresh")
}
func (m *TokenManager) parse(tokenString string, secret []byte, tokenType string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !token.Valid || claims.TokenType != tokenType {
		return nil, fmt.Errorf("invalid %s token", tokenType)
	}
	return claims, nil
}
