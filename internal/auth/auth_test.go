package auth

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestPasswordHash(t *testing.T) {
	h := NewHasher()
	ctx := context.Background()
	a, err := h.Hash(ctx, "a correct long password")
	if err != nil {
		t.Fatal(err)
	}
	b, err := h.Hash(ctx, "a correct long password")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("salt reused")
	}
	if err := h.Verify(ctx, "a correct long password", a); err != nil {
		t.Fatal(err)
	}
	for _, hash := range []string{a, "invalid", strings.Replace(a, "m=65536", "m=4294967295", 1), strings.Replace(a, "p=4", "p=0", 1)} {
		if err := h.Verify(ctx, "wrong password", hash); err == nil {
			t.Fatal("invalid password/hash accepted")
		}
	}
}
func TestLimiter(t *testing.T) {
	l := NewLimiter()
	now := time.Now()
	for i := 0; i < 2; i++ {
		if !l.allowAt(now, Limit{"ip", 2}, Limit{"account", 3}) {
			t.Fatal("early rejection")
		}
	}
	if l.allowAt(now, Limit{"ip", 2}, Limit{"account", 3}) {
		t.Fatal("limit bypassed")
	}
	if !l.allowAt(now.Add(16*time.Minute), Limit{"ip", 2}, Limit{"account", 3}) {
		t.Fatal("window did not reset")
	}
	if len(l.entries) != 2 {
		t.Fatal("expired entries not cleaned")
	}
	for i := 0; i < 10000; i++ {
		l.entries[string(rune(i))+"key"] = rateEntry{expires: now.Add(time.Hour)}
	}
	if l.allowAt(now, Limit{"new-key", 1}) {
		t.Fatal("unbounded limiter")
	}
}
func TestTokens(t *testing.T) {
	a, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b || len(a) != 43 || len(TokenHash(a)) != 64 || TokenHash(a) == CSRFToken(a) {
		t.Fatal("bad token")
	}
}
