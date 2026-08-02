package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/admin-template/backend/internal/pkg/auth"
)

func TestHashPassword_Success(t *testing.T) {
	hash, err := auth.HashPassword("correct horse battery staple")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "correct horse battery staple", hash, "hash must not echo plain text")
	// bcrypt hashes start with $2a$ / $2b$ / $2y$.
	assert.Regexp(t, `^\$2[aby]\$\d{2}\$`, hash)
}

func TestHashPassword_EmptyRejected(t *testing.T) {
	_, err := auth.HashPassword("")
	require.Error(t, err)
}

func TestVerifyPassword_Match(t *testing.T) {
	const plain = "s3cret-passphrase"
	hash, err := auth.HashPassword(plain)
	require.NoError(t, err)

	require.NoError(t, auth.VerifyPassword(hash, plain))
}

func TestVerifyPassword_Mismatch(t *testing.T) {
	hash, err := auth.HashPassword("correct")
	require.NoError(t, err)

	err = auth.VerifyPassword(hash, "wrong")
	require.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrInvalidCredentials),
		"mismatch must surface as ErrInvalidCredentials so the handler stays type-safe")
}

func TestVerifyPassword_EmptyRejected(t *testing.T) {
	hash, err := auth.HashPassword("nonempty")
	require.NoError(t, err)

	assert.ErrorIs(t, auth.VerifyPassword(hash, ""), auth.ErrInvalidCredentials)
	assert.ErrorIs(t, auth.VerifyPassword("", "nonempty"), auth.ErrInvalidCredentials)
	assert.ErrorIs(t, auth.VerifyPassword("", ""), auth.ErrInvalidCredentials)
}
