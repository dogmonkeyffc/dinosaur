package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"dinosaur/internal/history"
	"modernc.org/sqlite"
)

const historyColumns = `id,media_id,series_id,work_id,title,original_title,year,watched_on,rating,review,source,source_id,poster_key`

func scanHistory(row interface{ Scan(...any) error }) (history.Entry, error) {
	var e history.Entry
	err := row.Scan(&e.ID, &e.MediaID, &e.SeriesID, &e.WorkID, &e.Title, &e.OriginalTitle, &e.Year, &e.WatchedOn, &e.Rating, &e.Review, &e.Source, &e.SourceID, &e.PosterKey)
	return e, err
}

func (s *Store) HistoryByID(ctx context.Context, userID, id int64) (history.Entry, error) {
	return scanHistory(s.DB.QueryRowContext(ctx, `SELECT `+historyColumns+` FROM watch_history WHERE user_id=? AND id=?`, userID, id))
}

func (s *Store) HistoryPage(ctx context.Context, userID int64, page, size int) (history.Page, error) {
	result := history.Page{Items: []history.Entry{}, Page: page, PageSize: size}
	if page < 1 || size < 1 || size > 100 || page > 1000000 {
		return result, history.ErrInvalid
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM watch_history WHERE user_id=?`, userID).Scan(&result.Total); err != nil {
		return result, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT `+historyColumns+` FROM watch_history WHERE user_id=? ORDER BY watched_on DESC,id DESC LIMIT ? OFFSET ?`, userID, size, (page-1)*size)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		e, err := scanHistory(rows)
		if err != nil {
			rows.Close()
			return result, err
		}
		result.Items = append(result.Items, e)
	}
	err = rows.Err()
	if closeErr := rows.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return result, err
	}
	return result, tx.Commit()
}

const insertHistory = `INSERT INTO watch_history(user_id,media_id,series_id,work_id,title,original_title,year,watched_on,rating,review,source,source_id,poster_key,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

func historyArgs(userID int64, e history.Entry) []any {
	now := time.Now().Unix()
	return []any{userID, e.MediaID, e.SeriesID, e.WorkID, e.Title, e.OriginalTitle, e.Year, e.WatchedOn, e.Rating, e.Review, e.Source, e.SourceID, e.PosterKey, now, now}
}

func historyWriteError(err error) error {
	var se *sqlite.Error
	if errors.As(err, &se) && se.Code() == 2067 {
		return history.ErrConflict
	}
	return err
}

func (s *Store) CreateHistory(ctx context.Context, userID int64, e history.Entry) (int64, error) {
	if err := e.Validate(); err != nil {
		return 0, err
	}
	result, err := s.DB.ExecContext(ctx, insertHistory, historyArgs(userID, e)...)
	if err != nil {
		return 0, historyWriteError(err)
	}
	return result.LastInsertId()
}

func (s *Store) UpdateHistory(ctx context.Context, userID int64, e history.Entry) error {
	if err := e.Validate(); err != nil {
		return err
	}
	result, err := s.DB.ExecContext(ctx, `UPDATE watch_history SET title=?,original_title=?,year=?,watched_on=?,rating=?,review=?,updated_at=? WHERE user_id=? AND id=?`, e.Title, e.OriginalTitle, e.Year, e.WatchedOn, e.Rating, e.Review, time.Now().Unix(), userID, e.ID)
	if err != nil {
		return historyWriteError(err)
	}
	return historyAffected(result)
}

func (s *Store) DeleteHistory(ctx context.Context, userID, id int64) error {
	result, err := s.DB.ExecContext(ctx, `DELETE FROM watch_history WHERE user_id=? AND id=?`, userID, id)
	if err != nil {
		return err
	}
	return historyAffected(result)
}

func historyAffected(result sql.Result) error {
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (s *Store) ImportHistory(ctx context.Context, userID int64, entries []history.Entry) (history.ImportResult, error) {
	var result history.ImportResult
	for _, e := range entries {
		if err := e.Validate(); err != nil {
			return result, err
		}
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	for _, e := range entries {
		r, err := tx.ExecContext(ctx, insertHistory+` ON CONFLICT(user_id,source,source_id,watched_on) DO NOTHING`, historyArgs(userID, e)...)
		if err != nil {
			return history.ImportResult{}, err
		}
		n, err := r.RowsAffected()
		if err != nil {
			return history.ImportResult{}, err
		}
		if n == 0 {
			result.Skipped++
		} else {
			result.Imported++
		}
	}
	return result, tx.Commit()
}
