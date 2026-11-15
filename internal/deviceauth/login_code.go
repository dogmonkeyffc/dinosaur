package deviceauth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"dinosaur/internal/auth"
)

// Login codes share mobile grants so QR and numeric credentials replace each
// other for the same browser session. Only their hashes reach the repository.
func (s Service) CreateLoginCode(ctx context.Context, owner string) (Grant, string, error) {
	if owner == "" {
		return Grant{}, "", ErrInvalid
	}
	for attempt := 0; attempt < 5; attempt++ {
		number, err := rand.Int(rand.Reader, big.NewInt(100000000))
		if err != nil {
			return Grant{}, "", err
		}
		code := fmt.Sprintf("%08d", number)
		id, err := auth.NewToken()
		if err != nil {
			return Grant{}, "", err
		}
		grant := Grant{ID: id, Kind: "mobile", SecretHash: auth.TokenHash(code), OwnerHash: owner, ExpiresAt: time.Now().Add(2 * time.Minute).Unix()}
		err = s.Repository.CreateGrant(ctx, grant)
		if errors.Is(err, ErrConflict) {
			continue
		}
		return grant, code, err
	}
	return Grant{}, "", ErrConflict
}

func ValidLoginCode(code string) bool {
	if len(code) != 8 {
		return false
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func (s Service) RedeemLoginCode(ctx context.Context, code, name string) (Session, error) {
	if !ValidLoginCode(code) || !ValidName(name) {
		return Session{}, ErrInvalid
	}
	return s.redeem(ctx, "mobile", code, name)
}
