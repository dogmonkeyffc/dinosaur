package deviceauth

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"

	"dinosaur/internal/auth"
)

var (
	ErrExpired  = errors.New("pairing expired or already used")
	ErrPending  = errors.New("pairing awaiting approval")
	ErrConflict = errors.New("pairing already approved or code collision")
	ErrInvalid  = errors.New("invalid device name or pairing code")
	ErrFull     = errors.New("too many pending pairings")
)

type Grant struct {
	ID                          string `json:"pairing_id"`
	Code                        string `json:"code,omitempty"`
	Name                        string `json:"name"`
	ExpiresAt                   int64  `json:"expires_at"`
	SecretHash, OwnerHash, Kind string `json:"-"`
}

type Device struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
	Current   bool   `json:"current"`
}

type Session struct {
	Token     string    `json:"token"`
	User      auth.User `json:"user"`
	ExpiresAt int64     `json:"expires_at"`
}

type NewSession struct {
	Hash, ID, Name string
	Expires        time.Time
}

type Repository interface {
	CreateGrant(context.Context, Grant) error
	Pairing(context.Context, string) (Grant, error)
	ApprovePairing(context.Context, string, string) error
	RedeemGrant(context.Context, string, string, NewSession) (auth.User, error)
	Devices(context.Context, int64, string) ([]Device, error)
	RevokeDevice(context.Context, int64, string) error
	CancelGrant(context.Context, string) error
}

func ValidName(name string) bool {
	if strings.TrimSpace(name) != name || len(name) == 0 || len(name) > 128 {
		return false
	}
	for _, ch := range name {
		if unicode.IsControl(ch) {
			return false
		}
	}
	return true
}

func ValidCode(code string) bool {
	if len(code) != 6 {
		return false
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
