package deviceauth

import (
	"context"
	"dinosaur/internal/auth"
)

func (s Service) Cancel(ctx context.Context, secret string) error {
	if len(secret) != 43 {
		return ErrInvalid
	}
	return s.Repository.CancelGrant(ctx, auth.TokenHash(secret))
}

func (s Service) Lookup(ctx context.Context, code string) (Grant, error) {
	if !ValidCode(code) {
		return Grant{}, ErrInvalid
	}
	return s.Repository.Pairing(ctx, code)
}

func (s Service) Approve(ctx context.Context, id, owner string) error {
	if len(id) != 43 {
		return ErrInvalid
	}
	return s.Repository.ApprovePairing(ctx, id, owner)
}

func (s Service) List(ctx context.Context, userID int64, hash string) ([]Device, error) {
	return s.Repository.Devices(ctx, userID, hash)
}

func (s Service) Revoke(ctx context.Context, userID int64, id string) error {
	if len(id) != 43 {
		return ErrInvalid
	}
	return s.Repository.RevokeDevice(ctx, userID, id)
}
