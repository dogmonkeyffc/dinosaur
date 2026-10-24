package store

import (
	"context"
	"database/sql"
	"time"
)

func (s *Store) CreateSession(ctx context.Context, hash string, userID int64, passwordHash string, expires time.Time) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at<=?`, time.Now().Unix()); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO sessions(token_hash,user_id,expires_at,created_at) SELECT ?,id,?,? FROM users WHERE id=? AND disabled=0 AND password_hash=?`, hash, expires.Unix(), time.Now().Unix(), userID, passwordHash)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}
func (s *Store) SessionUser(ctx context.Context, hash string) (User, error) {
	return scanUser(s.DB.QueryRowContext(ctx, `SELECT u.id,u.username,u.password_hash,u.role,u.disabled,u.created_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=? AND s.expires_at>? AND u.disabled=0`, hash, time.Now().Unix()))
}
func (s *Store) DeleteSession(ctx context.Context, hash string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash=?`, hash)
	return err
}
