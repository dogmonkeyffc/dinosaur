package store

import (
	"context"
	"database/sql"
	"time"

	"dinosaur/internal/library"
)

func (s *Store) LibraryByID(ctx context.Context, id int64) (Library, error) {
	var item Library
	err := s.DB.QueryRowContext(ctx, `SELECT id,name,type,root_path,created_at,updated_at FROM libraries WHERE id=?`, id).Scan(&item.ID, &item.Name, &item.Type, &item.RootPath, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}
func (s *Store) NewScan(ctx context.Context, libraryID int64) (library.ScanJob, error) {
	job := library.ScanJob{LibraryID: libraryID, Status: "queued", CreatedAt: time.Now().Unix()}
	err := s.DB.QueryRowContext(ctx, `INSERT INTO scan_jobs(library_id,status,created_at) VALUES (?,'queued',?) RETURNING id`, libraryID, job.CreatedAt).Scan(&job.ID)
	return job, err
}
func (s *Store) SaveScan(ctx context.Context, job library.ScanJob) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE scan_jobs SET status=?,discovered=?,processed=?,errors=?,message=?,finished_at=? WHERE id=?`, job.Status, job.Discovered, job.Processed, job.Errors, job.Message, job.FinishedAt, job.ID)
	return err
}
func (s *Store) LatestScan(ctx context.Context, libraryID int64) (*library.ScanJob, error) {
	var job library.ScanJob
	err := s.DB.QueryRowContext(ctx, `SELECT id,library_id,status,discovered,processed,errors,message,created_at,finished_at FROM scan_jobs WHERE library_id=? ORDER BY id DESC LIMIT 1`, libraryID).Scan(&job.ID, &job.LibraryID, &job.Status, &job.Discovered, &job.Processed, &job.Errors, &job.Message, &job.CreatedAt, &job.FinishedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &job, err
}
func (s *Store) InterruptScans(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE scan_jobs SET status='interrupted',message='scan interrupted by server restart',finished_at=? WHERE status IN ('queued','scanning')`, time.Now().Unix())
	return err
}
