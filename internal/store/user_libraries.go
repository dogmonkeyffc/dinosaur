package store

import (
	"context"
	"database/sql"
	"errors"
)

var ErrLibraryUnrestricted = errors.New("media library is public")

type UserLibrary struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Restricted bool   `json:"restricted"`
	Granted    bool   `json:"granted"`
}

func (s *Store) UserLibraries(ctx context.Context, userID int64) ([]UserLibrary, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := libraryUser(ctx, tx, userID); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT l.id,l.name,l.restricted,g.user_id IS NOT NULL
 FROM libraries l LEFT JOIN library_grants g ON g.library_id=l.id AND g.user_id=?
 ORDER BY l.name COLLATE NOCASE,l.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []UserLibrary{}
	for rows.Next() {
		var item UserLibrary
		if err := rows.Scan(&item.ID, &item.Name, &item.Restricted, &item.Granted); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, tx.Commit()
}

func (s *Store) ChangeUserLibraries(ctx context.Context, userID int64, changes map[int64]bool) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := libraryUser(ctx, tx, userID); err != nil {
		return err
	}
	for id, granted := range changes {
		var restricted bool
		if err := tx.QueryRowContext(ctx, `SELECT restricted FROM libraries WHERE id=?`, id).Scan(&restricted); err != nil {
			return err
		}
		if !restricted {
			return ErrLibraryUnrestricted
		}
		if granted {
			_, err = tx.ExecContext(ctx, `INSERT INTO library_grants(library_id,user_id) VALUES(?,?) ON CONFLICT DO NOTHING`, id, userID)
		} else {
			_, err = tx.ExecContext(ctx, `DELETE FROM library_grants WHERE library_id=? AND user_id=?`, id, userID)
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func libraryUser(ctx context.Context, tx *sql.Tx, id int64) error {
	var role string
	if err := tx.QueryRowContext(ctx, `SELECT role FROM users WHERE id=?`, id).Scan(&role); err != nil {
		return err
	}
	if role != "user" {
		return ErrLibraryAccount
	}
	return nil
}
