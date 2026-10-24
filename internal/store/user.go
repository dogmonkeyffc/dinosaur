package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"modernc.org/sqlite"

	"dinosaur/internal/auth"
)

var (
	ErrInitialized = errors.New("already initialized")
	ErrConflict    = errors.New("username already exists")
	ErrLastAdmin   = errors.New("cannot disable the last active administrator")
)

type User = auth.User

func (s *Store) Initialized(ctx context.Context) (bool, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users)`).Scan(&n)
	return n != 0, err
}
func (s *Store) CreateUser(ctx context.Context, username, hash, role string, first bool) (User, error) {
	query := `INSERT INTO users(username,password_hash,role,created_at) VALUES (?,?,?,?) RETURNING id`
	if first {
		query = `INSERT INTO users(username,password_hash,role,created_at) SELECT ?,?,?,? WHERE NOT EXISTS (SELECT 1 FROM users) RETURNING id`
	}
	u := User{Username: username, PasswordHash: hash, Role: role, CreatedAt: time.Now().Unix()}
	err := s.DB.QueryRowContext(ctx, query, username, hash, role, u.CreatedAt).Scan(&u.ID)
	if first && errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrInitialized
	}
	var se *sqlite.Error
	if errors.As(err, &se) && se.Code() == 2067 {
		return User{}, ErrConflict
	}
	return u, err
}

const userColumns = `id,username,password_hash,role,disabled,created_at`

func scanUser(row interface{ Scan(...any) error }) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.Disabled, &u.CreatedAt)
	return u, err
}
func (s *Store) UserByName(ctx context.Context, name string) (User, error) {
	return scanUser(s.DB.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE username=?`, name))
}
func (s *Store) UserByID(ctx context.Context, id int64) (User, error) {
	return scanUser(s.DB.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id=?`, id))
}
func (s *Store) ChangePassword(ctx context.Context, id int64, hash string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, hash, id); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) SetDisabled(ctx context.Context, id int64, disabled bool) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	u, err := scanUser(tx.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id=?`, id))
	if err != nil {
		return err
	}
	if disabled && u.Role == "admin" && !u.Disabled {
		var n int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE role='admin' AND disabled=0`).Scan(&n); err != nil {
			return err
		}
		if n <= 1 {
			return ErrLastAdmin
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE users SET disabled=? WHERE id=?`, disabled, id); err != nil {
		return err
	}
	if disabled {
		if _, err = tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id=?`, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}
