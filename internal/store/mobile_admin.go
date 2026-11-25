package store

import (
	"context"
	"database/sql"
	"time"
)

func (s *Store) CreateMobileAdminGrant(ctx context.Context, parentHash, secretHash string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().Unix()
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash IN (SELECT token_hash FROM mobile_admin_sessions WHERE parent_hash=?)`, parentHash); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM mobile_admin_grants WHERE expires_at<=? OR parent_hash=?`, now, parentHash); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO mobile_admin_grants(secret_hash,parent_hash,expires_at)
 SELECT ?,s.token_hash,MIN(s.expires_at,?) FROM sessions s JOIN users u ON u.id=s.user_id
 WHERE s.token_hash=? AND s.kind='device' AND s.expires_at>? AND u.disabled=0 AND u.role='admin'`, secretHash, now+60, parentHash, now)
	if err != nil {
		return err
	}
	if err := historyAffected(result); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) CloseMobileAdminSessions(ctx context.Context, parentHash string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM mobile_admin_grants WHERE parent_hash=?`, parentHash); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash IN (SELECT token_hash FROM mobile_admin_sessions WHERE parent_hash=?)`, parentHash); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) RedeemMobileAdminGrant(ctx context.Context, secretHash, sessionHash string, expires time.Time) (time.Time, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}
	defer tx.Rollback()
	now := time.Now().Unix()
	var parent string
	var userID, parentExpiry int64
	err = tx.QueryRowContext(ctx, `SELECT p.token_hash,p.user_id,p.expires_at FROM mobile_admin_grants g
 JOIN sessions p ON p.token_hash=g.parent_hash JOIN users u ON u.id=p.user_id
 WHERE g.secret_hash=? AND g.expires_at>? AND p.kind='device' AND p.expires_at>? AND u.disabled=0 AND u.role='admin'`, secretHash, now, now).Scan(&parent, &userID, &parentExpiry)
	if err != nil {
		return time.Time{}, err
	}
	if expires.Unix() > parentExpiry {
		expires = time.Unix(parentExpiry, 0)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO sessions(token_hash,user_id,expires_at,created_at,kind) VALUES(?,?,?,?,'browser')`, sessionHash, userID, expires.Unix(), now); err != nil {
		return time.Time{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO mobile_admin_sessions VALUES(?,?)`, sessionHash, parent); err != nil {
		return time.Time{}, err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM mobile_admin_grants WHERE secret_hash=?`, secretHash)
	if err != nil {
		return time.Time{}, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return time.Time{}, err
	}
	if n != 1 {
		return time.Time{}, sql.ErrNoRows
	}
	return expires, tx.Commit()
}
