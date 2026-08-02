package repository

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/admin-template/backend/internal/model"
)

// RefreshTokenRepository manages the lifecycle of refresh tokens. The token
// hash is stored, not the raw token, so a DB leak does not let the attacker
// mint new access tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, userID uint64, raw string, expiresAt time.Time) error
	FindValid(ctx context.Context, raw string) (*model.RefreshToken, error)
	Revoke(ctx context.Context, id uint64) error
	RevokeAllForUser(ctx context.Context, userID uint64) error
	PurgeExpired(ctx context.Context) error
}

type refreshRepo struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshRepo{db: db}
}

// hashToken first reduces the JWT (200-500 bytes) to 32 bytes via SHA-256,
// then bcrypts the digest. bcrypt has a hard 72-byte input cap; the SHA-256
// prehash both sidesteps that and gives bcrypt a uniform-entropy input.
func hashToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

func (r *refreshRepo) Create(ctx context.Context, userID uint64, raw string, expiresAt time.Time) error {
	hash, err := bcrypt.GenerateFromPassword(hashToken(raw), bcrypt.MinCost)
	if err != nil {
		return err
	}
	rt := &model.RefreshToken{
		UserID:    userID,
		Hash:      string(hash),
		ExpiresAt: expiresAt,
	}
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *refreshRepo) FindValid(ctx context.Context, raw string) (*model.RefreshToken, error) {
	// Find candidates whose expiry is in the future and not revoked. Compare
	// the bcrypt hash in Go to keep the DB column a plain string.
	var candidates []model.RefreshToken
	err := r.db.WithContext(ctx).
		Where("expires_at > ? AND revoked = ?", time.Now(), false).
		Find(&candidates).Error
	if err != nil {
		return nil, err
	}
	digest := hashToken(raw)
	for i := range candidates {
		if err := bcrypt.CompareHashAndPassword([]byte(candidates[i].Hash), digest); err == nil {
			return &candidates[i], nil
		}
	}
	return nil, ErrNotFound
}

func (r *refreshRepo) Revoke(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("id = ?", id).Update("revoked", true).Error
}

func (r *refreshRepo) RevokeAllForUser(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("user_id = ?", userID).Update("revoked", true).Error
}

func (r *refreshRepo) PurgeExpired(ctx context.Context) error {
	res := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&model.RefreshToken{})
	if errors.Is(res.Error, gorm.ErrMissingWhereClause) {
		return nil // paranoia — Where clause is present
	}
	return res.Error
}
