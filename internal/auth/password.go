package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var ErrPassword = errors.New("invalid password")
var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,31}$`)

func NormalizeUsername(name string) string { return strings.ToLower(strings.TrimSpace(name)) }
func ValidateCredentials(name, password string) error {
	if !usernamePattern.MatchString(name) {
		return errors.New("username must be 3-32 ASCII letters, digits, dots, underscores or hyphens, starting with a letter or digit")
	}
	if len(password) < 12 || len(password) > 1024 {
		return errors.New("password must contain 12-1024 bytes")
	}
	return nil
}

// Hasher bounds expensive password operations to two concurrent allocations.
type Hasher struct{ slots chan struct{} }

func NewHasher() *Hasher { return &Hasher{slots: make(chan struct{}, 2)} }
func (h *Hasher) acquire(ctx context.Context) error {
	select {
	case h.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (h *Hasher) Hash(ctx context.Context, password string) (string, error) {
	if err := h.acquire(ctx); err != nil {
		return "", err
	}
	defer func() { <-h.slots }()
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=4$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}
func (h *Hasher) Verify(ctx context.Context, password, encoded string) error {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return ErrPassword
	}
	values := strings.Split(parts[3], ",")
	if len(values) != 3 {
		return ErrPassword
	}
	params := make([]uint64, 3)
	for i, prefix := range []string{"m=", "t=", "p="} {
		if !strings.HasPrefix(values[i], prefix) {
			return ErrPassword
		}
		n, err := strconv.ParseUint(strings.TrimPrefix(values[i], prefix), 10, 32)
		if err != nil {
			return ErrPassword
		}
		params[i] = n
	}
	if params[0] < 8192 || params[0] > 131072 || params[1] < 1 || params[1] > 5 || params[2] < 1 || params[2] > 8 {
		return ErrPassword
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		return ErrPassword
	}
	expected, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(expected) != 32 {
		return ErrPassword
	}
	if err := h.acquire(ctx); err != nil {
		return err
	}
	defer func() { <-h.slots }()
	key := argon2.IDKey([]byte(password), salt, uint32(params[1]), uint32(params[0]), uint8(params[2]), 32)
	if subtle.ConstantTimeCompare(key, expected) != 1 {
		return ErrPassword
	}
	return nil
}
func NewToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
func TokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func CSRFToken(token string) string { return TokenHash("dinosaur-csrf-v1:" + token) }
func Equal(a, b string) bool        { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }
