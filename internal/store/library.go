package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"dinosaur/internal/library"
	"modernc.org/sqlite"
)

var ErrLibraryPathExists = errors.New("media directory already belongs to a library")

type Library = library.Library

func (s *Store) Libraries(ctx context.Context) ([]Library, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,name,type,root_path,created_at,updated_at,(SELECT count(*) FROM media_items WHERE library_id=libraries.id AND missing=0) FROM libraries ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Library, 0)
	for rows.Next() {
		var item Library
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.RootPath, &item.CreatedAt, &item.UpdatedAt, &item.MediaCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (s *Store) SaveLibrary(ctx context.Context, item Library) (Library, error) {
	now := time.Now().Unix()
	var err error
	if item.ID == 0 {
		item.CreatedAt = now
		item.UpdatedAt = now
		err = s.DB.QueryRowContext(ctx, `INSERT INTO libraries(name,type,root_path,created_at,updated_at) VALUES (?,?,?,?,?) RETURNING id`, item.Name, item.Type, item.RootPath, now, now).Scan(&item.ID)
	} else {
		return s.updateLibrary(ctx, item)
	}
	var se *sqlite.Error
	if errors.As(err, &se) && se.Code() == 2067 {
		return Library{}, ErrLibraryPathExists
	}
	return item, err
}
func (s *Store) DeleteLibrary(ctx context.Context, id int64) error {
	result, err := s.DB.ExecContext(ctx, `DELETE FROM libraries WHERE id=?`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) updateLibrary(ctx context.Context, item Library) (Library, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return item, err
	}
	defer tx.Rollback()
	var oldRoot, oldType string
	err = tx.QueryRowContext(ctx, `SELECT root_path,type FROM libraries WHERE id=?`, item.ID).Scan(&oldRoot, &oldType)
	if err != nil {
		return item, err
	}
	item.UpdatedAt = time.Now().Unix()
	err = tx.QueryRowContext(ctx, `UPDATE libraries SET name=?,type=?,root_path=?,updated_at=? WHERE id=? RETURNING created_at`, item.Name, item.Type, item.RootPath, item.UpdatedAt, item.ID).Scan(&item.CreatedAt)
	var constraint *sqlite.Error
	if errors.As(err, &constraint) && constraint.Code() == 2067 {
		return item, ErrLibraryPathExists
	}
	if err != nil {
		return item, err
	}
	if oldRoot != item.RootPath || oldType != item.Type {
		if _, err := tx.ExecContext(ctx, `DELETE FROM media_items WHERE library_id=?`, item.ID); err != nil {
			return item, err
		}
	}
	return item, tx.Commit()
}
