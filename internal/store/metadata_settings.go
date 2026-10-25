package store

import (
	"context"
	"database/sql"

	"dinosaur/internal/metadata"
)

func readMetadataSettings(row *sql.Row) (metadata.Settings, error) {
	var settings metadata.Settings
	err := row.Scan(&settings.CredentialType, &settings.Language, &settings.Credential)
	return settings, err
}

func (s *Store) MetadataSettings(ctx context.Context) (metadata.Settings, error) {
	return readMetadataSettings(s.DB.QueryRowContext(ctx,
		`SELECT credential_type,language,credential FROM metadata_settings WHERE id=1`))
}

func (s *Store) UpdateMetadataSettings(ctx context.Context, update metadata.SettingsUpdate, clear bool) (metadata.Settings, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return metadata.Settings{}, err
	}
	defer tx.Rollback()
	settings, err := readMetadataSettings(tx.QueryRowContext(ctx,
		`SELECT credential_type,language,credential FROM metadata_settings WHERE id=1`))
	if err != nil {
		return metadata.Settings{}, err
	}
	settings, err = settings.Apply(update)
	if err != nil {
		return metadata.Settings{}, err
	}
	if clear {
		settings.Credential = ""
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE metadata_settings SET credential_type=?,language=?,credential=? WHERE id=1`,
		settings.CredentialType, settings.Language, settings.Credential); err != nil {
		return metadata.Settings{}, err
	}
	return settings, tx.Commit()
}
