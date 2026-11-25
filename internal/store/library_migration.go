package store

import (
	"context"
	"fmt"
	"time"
)

// Rebuild the parent table on one connection with FK enforcement restored before use.
// Turning enforcement off outside the transaction prevents DROP TABLE from cascading.
func (s *Store) migrateVideoLibraries(ctx context.Context) error {
	conn, err := s.DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	var exists int
	if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM schema_migrations WHERE version='024_video_libraries'`).Scan(&exists); err != nil {
		return err
	}
	if exists != 0 {
		return nil
	}
	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys=OFF`); err != nil {
		return err
	}
	defer conn.ExecContext(context.Background(), `PRAGMA foreign_keys=ON`)
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `CREATE TABLE libraries_video (
 id INTEGER PRIMARY KEY,
 name TEXT NOT NULL,
 type TEXT NOT NULL CHECK(type IN ('movie','tv','video')),
 root_path TEXT NOT NULL UNIQUE,
 created_at INTEGER NOT NULL,
 updated_at INTEGER NOT NULL,
 restricted INTEGER NOT NULL DEFAULT 0 CHECK(restricted IN (0,1))
 );
 INSERT INTO libraries_video SELECT id,name,type,root_path,created_at,updated_at,restricted FROM libraries;
 DROP TABLE libraries;
 ALTER TABLE libraries_video RENAME TO libraries;`)
	if err != nil {
		return fmt.Errorf("video library migration: %w", err)
	}
	rows, err := tx.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return err
	}
	invalid := rows.Next()
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if invalid {
		return fmt.Errorf("video library migration: foreign key check failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations VALUES('024_video_libraries',?)`, time.Now().Unix()); err != nil {
		return err
	}
	return tx.Commit()
}
