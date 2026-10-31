package store

import (
	"context"
	"database/sql"
	"time"

	"dinosaur/internal/agent"
)

const agentKeyColumns = `k.id,k.user_id,COALESCE(k.library_id,0),k.name,k.prefix,k.access,k.created_at,k.expires_at`

func scanAgentKey(row interface{ Scan(...any) error }) (agent.Key, error) {
	var key agent.Key
	err := row.Scan(&key.ID, &key.UserID, &key.LibraryID, &key.Name, &key.Prefix, &key.Access, &key.CreatedAt, &key.ExpiresAt)
	return key, err
}

func (s *Store) CreateAgentKey(ctx context.Context, key agent.Key, hash string) (agent.Key, error) {
	var library any
	if key.LibraryID > 0 {
		library = key.LibraryID
	}
	result, err := s.DB.ExecContext(ctx, `INSERT INTO agent_keys(user_id,library_id,name,token_hash,prefix,access,created_at,expires_at) VALUES (?,?,?,?,?,?,?,?)`, key.UserID, library, key.Name, hash, key.Prefix, key.Access, key.CreatedAt, key.ExpiresAt)
	if err != nil {
		return key, err
	}
	key.ID, err = result.LastInsertId()
	return key, err
}

func (s *Store) AgentKeys(ctx context.Context, userID int64) ([]agent.Key, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT `+agentKeyColumns+` FROM agent_keys k WHERE user_id=? ORDER BY k.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys := []agent.Key{}
	for rows.Next() {
		key, err := scanAgentKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (s *Store) AuthenticateAgentKey(ctx context.Context, hash string) (agent.Key, error) {
	return scanAgentKey(s.DB.QueryRowContext(ctx, `SELECT `+agentKeyColumns+` FROM agent_keys k JOIN users u ON u.id=k.user_id WHERE k.token_hash=? AND k.expires_at>? AND u.disabled=0 AND u.role='admin'`, hash, time.Now().Unix()))
}

func (s *Store) RevokeAgentKey(ctx context.Context, userID, id int64) error {
	result, err := s.DB.ExecContext(ctx, `DELETE FROM agent_keys WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return sql.ErrNoRows
	}
	return err
}
