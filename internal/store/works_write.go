package store

import (
	"context"
	"database/sql"
	"errors"

	"dinosaur/internal/television"
)

func (s *Store) SaveWork(ctx context.Context, id int64, title string, ids []int64) (television.Work, error) {
	result := television.Work{}
	title, err := television.ValidateWork(title, ids)
	if err != nil {
		return result, err
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	var libraryID int64
	if id != 0 {
		if err := tx.QueryRowContext(ctx, `SELECT library_id FROM works WHERE id=?`, id).Scan(&libraryID); err != nil {
			return result, err
		}
	}
	for _, memberID := range ids {
		var memberLibrary, memberWork int64
		err := tx.QueryRowContext(ctx, `SELECT library_id,COALESCE((SELECT work_id FROM work_members WHERE series_id=s.id),0) FROM tv_series s WHERE id=?`, memberID).Scan(&memberLibrary, &memberWork)
		if errors.Is(err, sql.ErrNoRows) {
			return result, &television.WorkError{Code: "series_not_found", Message: "series does not exist"}
		}
		if err != nil {
			return result, err
		}
		if err := television.ValidateWorkMember(libraryID, memberLibrary, id, memberWork); err != nil {
			return result, err
		}
		libraryID = memberLibrary
	}
	if id == 0 {
		if err := tx.QueryRowContext(ctx, `INSERT INTO works(title,library_id) VALUES(?,?) RETURNING id`, title, libraryID).Scan(&id); err != nil {
			return result, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE works SET title=? WHERE id=?`, title, id); err != nil {
			return result, err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM work_members WHERE work_id=?`, id); err != nil {
		return result, err
	}
	for position, memberID := range ids {
		if _, err := tx.ExecContext(ctx, `INSERT INTO work_members(work_id,series_id,position) VALUES(?,?,?)`, id, memberID, position); err != nil {
			return result, err
		}
	}
	if err := tx.Commit(); err != nil {
		return result, err
	}
	detail, err := s.WorkByID(ctx, id, 0)
	return detail.Work, err
}

func (s *Store) DeleteWork(ctx context.Context, id int64) error {
	result, err := s.DB.ExecContext(ctx, `DELETE FROM works WHERE id=?`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return sql.ErrNoRows
	}
	return err
}
