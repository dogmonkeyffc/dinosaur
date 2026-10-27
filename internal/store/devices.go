package store

import (
	"context"
	"time"

	"dinosaur/internal/deviceauth"
)

func (s *Store) Devices(ctx context.Context, userID int64, currentHash string) ([]deviceauth.Device, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT device_id,device_name,created_at,expires_at,token_hash=? FROM sessions WHERE user_id=? AND kind='device' AND expires_at>? ORDER BY created_at DESC,device_id`, currentHash, userID, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	devices := make([]deviceauth.Device, 0)
	for rows.Next() {
		var device deviceauth.Device
		if err := rows.Scan(&device.ID, &device.Name, &device.CreatedAt, &device.ExpiresAt, &device.Current); err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, rows.Err()
}

func (s *Store) RevokeDevice(ctx context.Context, userID int64, id string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE user_id=? AND device_id=? AND kind='device'`, userID, id)
	return err
}

func (s *Store) SessionUserKind(ctx context.Context, hash, kind string) (User, error) {
	return scanUser(s.DB.QueryRowContext(ctx, `SELECT u.id,u.username,u.password_hash,u.role,u.disabled,u.created_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=? AND s.kind=? AND s.expires_at>? AND u.disabled=0`, hash, kind, time.Now().Unix()))
}

func (s *Store) ServerID(ctx context.Context) (string, error) {
	var id string
	err := s.DB.QueryRowContext(ctx, `SELECT id FROM server_identity`).Scan(&id)
	return id, err
}
