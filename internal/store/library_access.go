package store

import (
	"context"
	"database/sql"
	"errors"
)

type libraryViewerKey struct{}

func WithLibraryViewer(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, libraryViewerKey{}, userID)
}

func libraryViewer(ctx context.Context) int64 {
	id, _ := ctx.Value(libraryViewerKey{}).(int64)
	return id
}

// libraryVisible takes a trusted SQL column name. Viewer zero is the
// unrestricted scope used by administrators and internal jobs.
func libraryVisible(column string) string {
	return `(NOT EXISTS(SELECT 1 FROM libraries acl WHERE acl.id=` + column + ` AND acl.restricted=1)
 OR ? IN (SELECT 0 UNION ALL SELECT g.user_id FROM library_grants g WHERE g.library_id=` + column + `))`
}

var ErrLibraryAccount = errors.New("invalid library account")

type LibraryAccess struct {
	Restricted bool    `json:"restricted"`
	UserIDs    []int64 `json:"user_ids"`
}

type LibraryAccount struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Disabled bool   `json:"disabled"`
}

func (s *Store) LibraryAccess(ctx context.Context, libraryID int64) (LibraryAccess, []LibraryAccount, error) {
	access := LibraryAccess{UserIDs: []int64{}}
	accounts := []LibraryAccount{}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return access, accounts, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, `SELECT restricted FROM libraries WHERE id=?`, libraryID).Scan(&access.Restricted); err != nil {
		return access, accounts, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT u.id,u.username,u.disabled,g.user_id IS NOT NULL
 FROM users u LEFT JOIN library_grants g ON g.user_id=u.id AND g.library_id=?
 WHERE u.role='user' ORDER BY u.username COLLATE NOCASE,u.id`, libraryID)
	if err != nil {
		return access, accounts, err
	}
	defer rows.Close()
	for rows.Next() {
		var account LibraryAccount
		var granted bool
		if err := rows.Scan(&account.ID, &account.Username, &account.Disabled, &granted); err != nil {
			return access, accounts, err
		}
		accounts = append(accounts, account)
		if granted {
			access.UserIDs = append(access.UserIDs, account.ID)
		}
	}
	if err := rows.Err(); err != nil {
		return access, accounts, err
	}
	return access, accounts, tx.Commit()
}

func (s *Store) SaveLibraryAccess(ctx context.Context, libraryID int64, access LibraryAccess) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE libraries SET restricted=? WHERE id=?`, access.Restricted, libraryID)
	if err != nil {
		return err
	}
	if count, err := result.RowsAffected(); err != nil {
		return err
	} else if count == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM library_grants WHERE library_id=?`, libraryID); err != nil {
		return err
	}
	for _, id := range access.UserIDs {
		var inserted int64
		err := tx.QueryRowContext(ctx, `INSERT INTO library_grants(library_id,user_id)
 SELECT ?,id FROM users WHERE id=? AND role='user' RETURNING user_id`, libraryID, id).Scan(&inserted)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLibraryAccount
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
