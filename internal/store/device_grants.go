package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"dinosaur/internal/auth"
	"dinosaur/internal/deviceauth"
	"modernc.org/sqlite"
)

func (s *Store) CreateGrant(ctx context.Context, grant deviceauth.Grant) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().Unix()
	if _, err := tx.ExecContext(ctx, `DELETE FROM device_grants WHERE expires_at<=?`, now); err != nil {
		return err
	}
	if grant.Kind == "mobile" {
		if _, err := tx.ExecContext(ctx, `DELETE FROM device_grants WHERE kind='mobile' AND owner_hash=?`, grant.OwnerHash); err != nil {
			return err
		}
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM device_grants`).Scan(&count); err != nil {
		return err
	}
	if count >= 1024 {
		return deviceauth.ErrFull
	}
	var code, owner any
	if grant.Code != "" {
		code = grant.Code
	}
	if grant.OwnerHash != "" {
		owner = grant.OwnerHash
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO device_grants(id,secret_hash,kind,code,name,owner_hash,expires_at) VALUES(?,?,?,?,?,?,?)`, grant.ID, grant.SecretHash, grant.Kind, code, grant.Name, owner, grant.ExpiresAt)
	var se *sqlite.Error
	if errors.As(err, &se) && se.Code() == 2067 {
		return deviceauth.ErrConflict
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) CancelGrant(ctx context.Context, hash string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM device_grants WHERE kind='tv' AND secret_hash=?`, hash)
	return err
}

func (s *Store) Pairing(ctx context.Context, code string) (deviceauth.Grant, error) {
	var grant deviceauth.Grant
	err := s.DB.QueryRowContext(ctx, `SELECT id,name,expires_at FROM device_grants WHERE code=? AND kind='tv' AND owner_hash IS NULL AND expires_at>?`, code, time.Now().Unix()).Scan(&grant.ID, &grant.Name, &grant.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return grant, deviceauth.ErrExpired
	}
	return grant, err
}

func (s *Store) ApprovePairing(ctx context.Context, id, owner string) error {
	now := time.Now().Unix()
	result, err := s.DB.ExecContext(ctx, `UPDATE device_grants SET owner_hash=? WHERE id=? AND kind='tv' AND owner_hash IS NULL AND expires_at>? AND EXISTS(SELECT 1 FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=? AND s.expires_at>? AND u.disabled=0)`, owner, id, now, owner, now)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n != 1 {
		return deviceauth.ErrExpired
	}
	return err
}

func (s *Store) RedeemGrant(ctx context.Context, kind, secret string, session deviceauth.NewSession) (auth.User, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return auth.User{}, err
	}
	defer tx.Rollback()
	now := time.Now().Unix()
	var id, name string
	var owner sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT id,name,owner_hash FROM device_grants WHERE secret_hash=? AND kind=? AND expires_at>?`, secret, kind, now).Scan(&id, &name, &owner)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, deviceauth.ErrExpired
	}
	if err != nil {
		return auth.User{}, err
	}
	if !owner.Valid {
		return auth.User{}, deviceauth.ErrPending
	}
	user, err := scanUser(tx.QueryRowContext(ctx, `SELECT u.id,u.username,u.password_hash,u.role,u.disabled,u.created_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=? AND s.expires_at>? AND u.disabled=0`, owner.String, now))
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, deviceauth.ErrExpired
	}
	if err != nil {
		return auth.User{}, err
	}
	if kind == "mobile" {
		name = session.Name
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO sessions(token_hash,user_id,expires_at,created_at,kind,device_id,device_name) VALUES(?,?,?,?,'device',?,?)`, session.Hash, user.ID, session.Expires.Unix(), now, session.ID, name); err != nil {
		return auth.User{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM device_grants WHERE id=?`, id); err != nil {
		return auth.User{}, err
	}
	if err := tx.Commit(); err != nil {
		return auth.User{}, err
	}
	return user, nil
}
