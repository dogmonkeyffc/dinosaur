package store

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"dinosaur/internal/auth"
	"dinosaur/internal/deviceauth"
)

func loginCodeStore(t *testing.T) (*Store, deviceauth.Service, User) {
	t.Helper()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "codes.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	user, err := db.CreateUser(ctx, "viewer", "hash", "user", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSession(ctx, "owner", user.ID, user.PasswordHash, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	return db, deviceauth.Service{Repository: db, TTL: time.Hour}, user
}

func TestLoginCodeConcurrentSingleUse(t *testing.T) {
	_, service, user := loginCodeStore(t)
	ctx := context.Background()
	_, code, err := service.CreateLoginCode(ctx, "owner")
	if err != nil {
		t.Fatal(err)
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Go(func() {
			session, err := service.RedeemLoginCode(ctx, code, "phone")
			if err == nil {
				successes.Add(1)
				if session.User.ID != user.ID {
					t.Error("wrong user")
				}
			} else if !errors.Is(err, deviceauth.ErrExpired) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successes=%d", successes.Load())
	}
}

func TestLoginCodeTransactionFailureDoesNotConsume(t *testing.T) {
	db, service, _ := loginCodeStore(t)
	ctx := context.Background()
	_, code, err := service.CreateLoginCode(ctx, "owner")
	if err != nil {
		t.Fatal(err)
	}
	// Fail after session insertion but before grant deletion: both must roll back.
	if _, err := db.DB.Exec(`CREATE TRIGGER fail_redeem BEFORE DELETE ON device_grants BEGIN SELECT RAISE(ABORT, 'injected failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RedeemLoginCode(ctx, code, "phone"); err == nil {
		t.Fatal("expected transaction failure")
	}
	var count int
	if err := db.DB.QueryRow(`SELECT count(*) FROM sessions WHERE kind='device'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial device session count=%d err=%v", count, err)
	}
	if _, err := db.DB.Exec(`DROP TRIGGER fail_redeem`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RedeemLoginCode(ctx, code, "phone"); err != nil {
		t.Fatalf("failed transaction consumed code: %v", err)
	}
}

func TestLoginCodeOwnerRevocation(t *testing.T) {
	for _, operation := range []string{"password", "disabled", "expired", "logout"} {
		t.Run(operation, func(t *testing.T) {
			db, service, user := loginCodeStore(t)
			ctx := context.Background()
			_, code, err := service.CreateLoginCode(ctx, "owner")
			if err != nil {
				t.Fatal(err)
			}
			switch operation {
			case "password":
				err = db.ChangePassword(ctx, user.ID, "new-hash")
			case "disabled":
				_, err = db.DB.Exec(`UPDATE users SET disabled=1 WHERE id=?`, user.ID)
			case "expired":
				_, err = db.DB.Exec(`UPDATE sessions SET expires_at=?`, time.Now().Unix())
			case "logout":
				err = db.DeleteSession(ctx, "owner")
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := service.RedeemLoginCode(ctx, code, "phone"); !errors.Is(err, deviceauth.ErrExpired) {
				t.Fatalf("revoked owner redeemed: %v", err)
			}
		})
	}
}

func TestLoginCodeCollisionAndReplacementRollback(t *testing.T) {
	db, service, user := loginCodeStore(t)
	ctx := context.Background()
	_, old, err := service.CreateLoginCode(ctx, "owner")
	if err != nil {
		t.Fatal(err)
	}
	sameOwner := deviceauth.Grant{ID: "same-owner", Kind: "mobile", OwnerHash: "owner", SecretHash: auth.TokenHash(old), ExpiresAt: time.Now().Add(time.Minute).Unix()}
	if err := db.CreateGrant(ctx, sameOwner); !errors.Is(err, deviceauth.ErrConflict) {
		t.Fatalf("same-owner collision not retried: %v", err)
	}
	if err := db.CreateSession(ctx, "second", user.ID, user.PasswordHash, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	other, _, err := service.CreateLoginCode(ctx, "second")
	if err != nil {
		t.Fatal(err)
	}
	collision := deviceauth.Grant{ID: "collision", Kind: "mobile", OwnerHash: "owner", SecretHash: other.SecretHash, ExpiresAt: time.Now().Add(time.Minute).Unix()}
	if err := db.CreateGrant(ctx, collision); !errors.Is(err, deviceauth.ErrConflict) {
		t.Fatalf("collision not detected: %v", err)
	}
	if _, err := service.RedeemLoginCode(ctx, old, "phone"); err != nil {
		t.Fatalf("collision consumed previous owner code: %v", err)
	}
	// A device session cannot become the root of a new mobile authorization.
	native, err := service.RedeemLoginCode(ctx, mustCode(t, service, "second"), "phone")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.CreateLoginCode(ctx, auth.TokenHash(native.Token)); !errors.Is(err, deviceauth.ErrExpired) {
		t.Fatalf("device owner accepted: %v", err)
	}
}

func mustCode(t *testing.T, service deviceauth.Service, owner string) string {
	t.Helper()
	_, code, err := service.CreateLoginCode(context.Background(), owner)
	if err != nil {
		t.Fatal(err)
	}
	return code
}
