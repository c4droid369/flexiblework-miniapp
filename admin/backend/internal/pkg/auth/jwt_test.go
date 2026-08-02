package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/admin-template/backend/internal/pkg/auth"
)

const (
	testSecret = "this-is-a-32-byte-test-secret-ok"
)

func newIssuer(t *testing.T) *auth.Issuer {
	t.Helper()
	iss, err := auth.NewIssuer(testSecret, time.Hour, 24*time.Hour, "test-iss")
	require.NoError(t, err)
	return iss
}

func TestNewIssuer_SecretTooShort(t *testing.T) {
	_, err := auth.NewIssuer("too-short", time.Hour, time.Hour, "x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "32 bytes")
}

func TestNewIssuer_NonPositiveTTLError(t *testing.T) {
	_, err := auth.NewIssuer(testSecret, 0, time.Hour, "x")
	require.Error(t, err)
	_, err = auth.NewIssuer(testSecret, time.Hour, 0, "x")
	require.Error(t, err)
}

func TestIssueAccess_Parse_RoundTrip(t *testing.T) {
	iss := newIssuer(t)
	tok, exp, err := iss.IssueAccess(42, []string{"user:view", "user:create"})
	require.NoError(t, err)
	require.NotEmpty(t, tok)
	assert.True(t, exp.After(time.Now()), "expiry must be in the future")

	claims, err := iss.Parse(tok, auth.KindAccess)
	require.NoError(t, err)
	assert.Equal(t, uint64(42), claims.UserID)
	assert.Equal(t, auth.KindAccess, claims.Kind)
	assert.Equal(t, []string{"user:view", "user:create"}, claims.Permissions)
	assert.Equal(t, "test-iss", claims.Issuer)
	assert.Equal(t, "42", claims.Subject)
}

func TestParse_WrongKind(t *testing.T) {
	iss := newIssuer(t)
	tok, _, err := iss.IssueRefresh(7)
	require.NoError(t, err)

	_, err = iss.Parse(tok, auth.KindAccess)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kind")
}

func TestParse_Expired(t *testing.T) {
	// Manually mint a token whose exp is in the past — bypasses NewIssuer's
	// positive-TTL check.
	now := time.Now().Add(-time.Hour)
	claims := auth.Claims{
		Kind:   auth.KindAccess,
		UserID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-iss",
			IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(now),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(testSecret))
	require.NoError(t, err)

	iss := newIssuer(t)
	_, err = iss.Parse(signed, auth.KindAccess)
	require.Error(t, err)
}

func TestParse_TamperedSignature(t *testing.T) {
	iss := newIssuer(t)
	tok, _, err := iss.IssueAccess(1, nil)
	require.NoError(t, err)

	// Flip a char in the middle of the signature segment (chars 60..end).
	// Pick a position guaranteed to be in the signature, not the payload.
	sigStart := len(tok) - 10
	tampered := tok[:sigStart] + flipChar(tok[sigStart])
	require.NotEqual(t, tok, tampered)

	_, err = iss.Parse(tampered, auth.KindAccess)
	require.Error(t, err, "tampered signature must be rejected")
}

// flipChar returns the next printable base64-url char. base64url alphabet
// is [A-Za-z0-9_-]; we just need something that decodes to a different byte.
func flipChar(b byte) string {
	switch {
	case b >= 'A' && b < 'Z':
		return string(b + 1)
	case b == 'Z':
		return "a"
	default:
		return "A"
	}
}

func TestParse_AlgConfusion(t *testing.T) {
	iss := newIssuer(t)
	// Hand-craft an unsigned (alg=none) JWT to verify Parse rejects it.
	none := jwt.NewWithClaims(jwt.SigningMethodNone, &auth.Claims{Kind: auth.KindAccess, UserID: 1})
	raw, err := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = iss.Parse(raw, auth.KindAccess)
	require.Error(t, err, "unsigned tokens must be rejected")
}

func TestUserIDFrom_EmptyContext(t *testing.T) {
	assert.Equal(t, uint64(0), auth.UserIDFrom(context.Background()))
}

func TestWithUserID_RoundTrip(t *testing.T) {
	ctx := auth.WithUserID(context.Background(), 99)
	assert.Equal(t, uint64(99), auth.UserIDFrom(ctx))
}
