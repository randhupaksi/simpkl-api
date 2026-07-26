package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTokenManagerIssuesTypedTokens(t *testing.T) {
	manager := NewTokenManager(
		"access-secret-long-enough",
		"refresh-secret-long-enough",
		time.Minute,
		time.Hour,
	)
	access, refresh, err := manager.Issue("user-id", "staff@example.sch.id")
	require.NoError(t, err)
	accessClaims, err := manager.ParseAccess(access)
	require.NoError(t, err)
	require.Equal(t, "user-id", accessClaims.UserID)
	require.Equal(t, "access", accessClaims.TokenType)
	refreshClaims, err := manager.ParseRefresh(refresh)
	require.NoError(t, err)
	require.Equal(t, "refresh", refreshClaims.TokenType)
	_, err = manager.ParseAccess(refresh)
	require.Error(t, err)
}

func TestPasswordHash(t *testing.T) {
	hash, err := HashPassword("strong-password")
	require.NoError(t, err)
	require.NoError(t, ComparePassword(hash, "strong-password"))
	require.Error(t, ComparePassword(hash, "wrong-password"))
}
