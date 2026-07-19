//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAccountConcurrencyDefaultsInvalidGrokOAuthToOne(t *testing.T) {
	require.Equal(t, 1, normalizeAccountConcurrency(PlatformGrok, AccountTypeOAuth, 0))
	require.Equal(t, 1, normalizeAccountConcurrency(PlatformGrok, AccountTypeOAuth, -5))
}

func TestNormalizeAccountConcurrencyPreservesExplicitValues(t *testing.T) {
	require.Equal(t, 50, normalizeAccountConcurrency(PlatformGrok, AccountTypeOAuth, 50))
	require.Equal(t, 2, normalizeAccountConcurrency(PlatformOpenAI, AccountTypeOAuth, 2))
	require.Equal(t, 2, normalizeAccountConcurrency(PlatformGrok, AccountTypeAPIKey, 2))
}

func TestCreateAccountNormalizesLegacyXAIPlatform(t *testing.T) {
	repo := &mockAccountRepoForPlatform{}
	svc := &adminServiceImpl{accountRepo: repo}

	account, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:                 "legacy-xai-oauth",
		Platform:             " XAI ",
		Type:                 AccountTypeOAuth,
		Credentials:          map[string]any{"access_token": "test-token"},
		SkipDefaultGroupBind: true,
	})
	require.NoError(t, err)
	require.Equal(t, PlatformGrok, account.Platform)
	require.Equal(t, 1, account.Concurrency)
}
