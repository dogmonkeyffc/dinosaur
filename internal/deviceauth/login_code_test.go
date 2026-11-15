package deviceauth

import (
	"context"
	"errors"
	"testing"
	"time"

	"dinosaur/internal/auth"
)

type collisionRepository struct {
	Repository
	calls      int
	collisions int
	grant      Grant
}

func (r *collisionRepository) CreateGrant(_ context.Context, grant Grant) error {
	r.calls++
	if r.calls <= r.collisions {
		return ErrConflict
	}
	r.grant = grant
	return nil
}

func TestLoginCodeGenerationRetriesCollisions(t *testing.T) {
	repository := &collisionRepository{collisions: 2}
	service := Service{Repository: repository}
	grant, code, err := service.CreateLoginCode(context.Background(), "owner")
	if err != nil || repository.calls != 3 {
		t.Fatalf("calls=%d err=%v", repository.calls, err)
	}
	if !ValidLoginCode(code) || grant.Code != "" || grant.SecretHash != auth.TokenHash(code) || grant.OwnerHash != "owner" || grant.Kind != "mobile" || grant.ExpiresAt < time.Now().Unix()+119 {
		t.Fatalf("invalid generated grant: %+v", grant)
	}
	repository = &collisionRepository{collisions: 10}
	service.Repository = repository
	if _, _, err := service.CreateLoginCode(context.Background(), "owner"); !errors.Is(err, ErrConflict) || repository.calls != 5 {
		t.Fatalf("retry exhaustion calls=%d err=%v", repository.calls, err)
	}
}

func TestLoginCodeValidationDoesNotCallRepository(t *testing.T) {
	service := Service{}
	if _, _, err := service.CreateLoginCode(context.Background(), ""); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	for _, code := range []string{"123456", "123456789", "1234567x", "１２３４５６７８"} {
		if _, err := service.RedeemLoginCode(context.Background(), code, "phone"); !errors.Is(err, ErrInvalid) {
			t.Fatal(err)
		}
	}
	if _, err := service.RedeemLoginCode(context.Background(), "12345678", "\n"); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := service.Redeem(context.Background(), "mobile", "12345678", "phone"); !errors.Is(err, ErrInvalid) {
		t.Fatal("numeric code accepted as QR ticket")
	}
	if _, err := service.Lookup(context.Background(), "12345678"); !errors.Is(err, ErrInvalid) {
		t.Fatal("mobile code accepted as TV code")
	}
}
