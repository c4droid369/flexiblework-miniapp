package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Kind discriminates the token type. Refresh tokens carry the same shape as
// access tokens but a longer TTL and a "refresh" kind; middleware uses Kind
// to reject mismatched use.
type Kind string

const (
	KindAccess  Kind = "access"
	KindRefresh Kind = "refresh"
)

// Claims is the JWT body. Permissions are included on issue so that
// downstream middleware can do an O(1) check without hitting the DB on every
// request. Stale permissions last until access-token expiry (default 2h) —
// acceptable for an admin template; rotate via /auth/refresh to pick up
// changes sooner.
type Claims struct {
	Kind        Kind     `json:"kind"`
	UserID      uint64   `json:"uid"`
	Permissions []string `json:"perms,omitempty"`
	jwt.RegisteredClaims
}

// Issuer mints and parses access/refresh tokens. It is safe for concurrent
// use by multiple goroutines.
type Issuer struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
}

// NewIssuer validates the secret length and returns a configured Issuer.
// Secret must be at least 32 bytes (256 bits) — anything shorter is rejected
// to prevent HS256 weak-key attacks.
func NewIssuer(secret string, accessTTL, refreshTTL time.Duration, issuer string) (*Issuer, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("jwt secret must be ≥ 32 bytes (got %d)", len(secret))
	}
	if accessTTL <= 0 || refreshTTL <= 0 {
		return nil, errors.New("jwt ttls must be positive")
	}
	return &Issuer{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		issuer:     issuer,
	}, nil
}

// IssueAccess mints a fresh access token bound to userID with the supplied
// permission codes baked into the claims.
func (i *Issuer) IssueAccess(userID uint64, permissions []string) (string, time.Time, error) {
	return i.issue(userID, KindAccess, i.accessTTL, permissions)
}

// IssueRefresh mints a fresh refresh token. The token is a JWT, but the
// service layer must persist its hash to allow revocation. Permissions are
// intentionally NOT included — refresh tokens are used only at /auth/refresh
// to mint a fresh access token; no permission checks happen on them.
func (i *Issuer) IssueRefresh(userID uint64) (string, time.Time, error) {
	return i.issue(userID, KindRefresh, i.refreshTTL, nil)
}

func (i *Issuer) issue(userID uint64, kind Kind, ttl time.Duration, perms []string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(ttl)
	claims := Claims{
		Kind:        kind,
		UserID:      userID,
		Permissions: perms,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        uuid.Must(uuid.NewV7()).String(),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// Parse validates signature, expiry, and kind. Returns Claims on success.
func (i *Issuer) Parse(raw string, expected Kind) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return i.secret, nil
	}, jwt.WithIssuer(i.issuer), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Kind != expected {
		return nil, fmt.Errorf("unexpected token kind: %s", claims.Kind)
	}
	return claims, nil
}

// ctxKey is unexported so other packages cannot collide on context values.
type ctxKey struct{ name string }

var userIDKey = ctxKey{"auth.user_id"}

// WithUserID returns a child context carrying the authenticated user's id.
func WithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFrom extracts the user id set by auth middleware; 0 if absent.
func UserIDFrom(ctx context.Context) uint64 {
	if v, ok := ctx.Value(userIDKey).(uint64); ok {
		return v
	}
	return 0
}
