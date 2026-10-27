package store

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"dinosaur/internal/deviceauth"
)

func TestConcurrentGrantRedemptionAndPersistentIdentity(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "devices.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	user, err := db.CreateUser(ctx, "viewer", "password-hash", "user", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSession(ctx, "owner", user.ID, user.PasswordHash, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	service := deviceauth.Service{Repository: db, TTL: time.Hour}
	_, ticket, err := service.Create(ctx, "mobile", "owner", "")
	if err != nil {
		t.Fatal(err)
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Go(func() {
			session, err := service.Redeem(ctx, "mobile", ticket, "手机")
			if err == nil {
				successes.Add(1)
				if session.User.ID != user.ID {
					t.Error("wrong account")
				}
			} else if !errors.Is(err, deviceauth.ErrExpired) {
				t.Errorf("unexpected redemption error: %v", err)
			}
		})
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful redemptions=%d", successes.Load())
	}
	devices, err := db.Devices(ctx, user.ID, "")
	if err != nil || len(devices) != 1 {
		t.Fatalf("devices=%d error=%v", len(devices), err)
	}
	var raw string
	if err := db.DB.QueryRow(`SELECT token_hash FROM sessions WHERE kind='device'`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	if len(raw) != 64 {
		t.Fatal("device token is not stored as a hash")
	}
	identity, err := db.ServerID(ctx)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	identity2, err := second.ServerID(ctx)
	if err != nil || identity != identity2 || len(identity) != 32 {
		t.Fatal("server identity changed")
	}
}

func TestPasswordChangeRevokesDevicesAndPendingApprovals(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "devices.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	user, err := db.CreateUser(ctx, "viewer", "hash", "user", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSession(ctx, "owner", user.ID, user.PasswordHash, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	service := deviceauth.Service{Repository: db, TTL: time.Hour}
	qr, secret, err := service.Create(ctx, "mobile", "owner", "")
	if err != nil || qr.ExpiresAt <= time.Now().Unix() {
		t.Fatal("QR creation failed")
	}
	if _, err := service.Redeem(ctx, "mobile", secret, "手机"); err != nil {
		t.Fatal(err)
	}
	grant, secret, err := service.Create(ctx, "tv", "", "电视")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Approve(ctx, grant.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	if err := db.ChangePassword(ctx, user.ID, "new-hash"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Redeem(ctx, "tv", secret, ""); !errors.Is(err, deviceauth.ErrExpired) {
		t.Fatalf("pending TV login survived password change: %v", err)
	}
	devices, err := service.List(ctx, user.ID, "")
	if err != nil || len(devices) != 0 {
		t.Fatal("device survived password change")
	}
}
