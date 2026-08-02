// Package auth owns authentication primitives: bcrypt hashing and JWT
// issue/parse. Higher layers (service, handlers, middleware) compose these.
package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the default work factor. 12 ≈ 250ms on modern hardware.
	BcryptCost = 12
)

// ErrInvalidCredentials signals a bad username/password combination. Services
// return it to keep the cause type-checked at the handler boundary.
var ErrInvalidCredentials = errors.New("invalid credentials")

// HashPassword returns a bcrypt hash of plain. Empty plain is rejected.
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("empty password")
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// VerifyPassword compares a bcrypt hash with the candidate plain text. Use
// the returned error to distinguish "wrong password" (ErrInvalidCredentials)
// from internal failure.
func VerifyPassword(hash, plain string) error {
	if hash == "" || plain == "" {
		return ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}
