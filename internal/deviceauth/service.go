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

type Service struct {
	Repository Repository
	TTL        time.Duration
}

func (s Service) Create(ctx context.Context, kind, owner, name string) (Grant, string, error) {
	if kind != "mobile" && kind != "tv" || kind == "tv" && !ValidName(name) || kind == "mobile" && owner == "" {
		return Grant{}, "", ErrInvalid
	}
	for attempt := 0; attempt < 5; attempt++ {
		secret, err := auth.NewToken()
		if err != nil {
			return Grant{}, "", err
		}
		id, err := auth.NewToken()
		if err != nil {
			return Grant{}, "", err
		}
		grant := Grant{ID: id, Name: name, Kind: kind, SecretHash: auth.TokenHash(secret), OwnerHash: owner}
		lifetime := 2 * time.Minute
		if kind == "tv" {
			code, err := rand.Int(rand.Reader, big.NewInt(1000000))
			if err != nil {
				return Grant{}, "", err
			}
			grant.Code = fmt.Sprintf("%06d", code)
			lifetime = 5 * time.Minute
		}
		grant.ExpiresAt = time.Now().Add(lifetime).Unix()
		err = s.Repository.CreateGrant(ctx, grant)
		if errors.Is(err, ErrConflict) {
			continue
		}
		return grant, secret, err
	}
	return Grant{}, "", ErrConflict
}

func (s Service) Redeem(ctx context.Context, kind, secret, name string) (Session, error) {
	if len(secret) != 43 || kind != "mobile" && kind != "tv" || kind == "mobile" && !ValidName(name) {
		return Session{}, ErrInvalid
	}
	token, err := auth.NewToken()
	if err != nil {
		return Session{}, err
	}
	id, err := auth.NewToken()
	if err != nil {
		return Session{}, err
	}
	expires := time.Now().Add(s.TTL)
	user, err := s.Repository.RedeemGrant(ctx, kind, auth.TokenHash(secret), NewSession{Hash: auth.TokenHash(token), ID: id, Name: name, Expires: expires})
	if err != nil {
		return Session{}, err
	}
	return Session{Token: token, User: user, ExpiresAt: expires.Unix()}, nil
}
