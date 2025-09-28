package x3dh

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type KeyStore interface {
	GetIdentityKey(context.Context, string) (*IdentityKey, error)
	GetSignedPreKey(context.Context, string) (*SignedPreKey, error)
	ConsumeOneTimePreKey(context.Context, string) (*OneTimePreKey, error)
	CountOPKs(context.Context, string) (int, error)
	StoreBundle(context.Context, RegisterBundleRequest, CryptoService) error
	InsertOPKs(context.Context, string, []OneTimePreKeyUpload) error
	CountAvailableOPKs(context.Context, string) (int, error)
	ListUsersWithOPKs(context.Context) ([]string, error)
	RotateSignedPreKey(context.Context, string, string, string, *time.Time) error
}

var ErrNoOPKsAvailable = errors.New("no available one-time prekeys")
var _ KeyStore = (*PostgresKeyStore)(nil)

type PostgresKeyStore struct {
	DB *sql.DB
}

func NewPostgresKeyStore(db *sql.DB) *PostgresKeyStore {
	return &PostgresKeyStore{DB: db}
}
func (s *PostgresKeyStore) GetIdentityKey(ctx context.Context, userID string) (*IdentityKey, error) {
	row := s.DB.QueryRowContext(ctx, `
        SELECT public_key
        FROM x3dh_identity_keys
        WHERE user_id = $1
    `, userID)

	var ik IdentityKey
	ik.UserID = userID
	if err := row.Scan(&ik.PublicKey); err != nil {
		return nil, fmt.Errorf("failed to fetch identity key: %w", err)
	}
	return &ik, nil
}

func (s *PostgresKeyStore) GetSignedPreKey(ctx context.Context, userID string) (*SignedPreKey, error) {
	row := s.DB.QueryRowContext(ctx, `
        SELECT public_key, signature
        FROM x3dh_signed_prekeys
        WHERE user_id = $1
    `, userID)

	var spk SignedPreKey
	spk.UserID = userID
	if err := row.Scan(&spk.PublicKey, &spk.Signature); err != nil {
		return nil, fmt.Errorf("failed to fetch signed prekey: %w", err)
	}
	return &spk, nil
}

func (s *PostgresKeyStore) ConsumeOneTimePreKey(ctx context.Context, userID string) (*OneTimePreKey, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
        SELECT id, key_id, public_key
        FROM x3dh_one_time_prekeys
        WHERE user_id = $1 AND is_used = FALSE
        ORDER BY created_at ASC
        LIMIT 1
        FOR UPDATE
    `, userID)

	var opk OneTimePreKey
	opk.UserID = userID
	if err := row.Scan(&opk.ID, &opk.KeyID, &opk.PublicKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoOPKsAvailable
		}
		return nil, fmt.Errorf("failed to fetch one-time prekey: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
        UPDATE x3dh_one_time_prekeys SET is_used = TRUE WHERE id = $1
    `, opk.ID); err != nil {
		return nil, fmt.Errorf("failed to mark one-time prekey as used: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// ✅ return the scanned struct, not undefined vars
	return &opk, nil
}

// keystore_postgres.go
func (s *PostgresKeyStore) StoreBundle(ctx context.Context, req RegisterBundleRequest, _ CryptoService) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
        INSERT INTO x3dh_identity_keys (user_id, public_key)
        VALUES ($1, $2)
        ON CONFLICT (user_id) DO UPDATE SET public_key = EXCLUDED.public_key
    `, req.UserID, req.IdentityKey)
	if err != nil {
		return fmt.Errorf("failed to insert IK: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO x3dh_signed_prekeys (user_id, public_key, signature, created_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (user_id) DO UPDATE 
            SET public_key = EXCLUDED.public_key,
                signature = EXCLUDED.signature,
                created_at = EXCLUDED.created_at
    `, req.UserID, req.SignedPreKey, req.SPKSignature)
	if err != nil {
		return fmt.Errorf("failed to insert SPK: %w", err)
	}

	for _, opk := range req.OneTimePreKeys {
		_, err := tx.ExecContext(ctx, `
        INSERT INTO x3dh_one_time_prekeys (user_id, key_id, public_key)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, key_id) DO NOTHING
    `, req.UserID, opk.KeyID, opk.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to insert OPK: %w", err)
		}
	}

	return tx.Commit()
}

func (s *PostgresKeyStore) CountOPKs(ctx context.Context, userID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM x3dh_one_time_prekeys 
		WHERE user_id = $1 AND is_used = false
	`
	err := s.DB.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

// in internal/X3dh/keystore_postgres.go (same file as other PostgresKeyStore methods)
func (s *PostgresKeyStore) CountAvailableOPKs(ctx context.Context, userID string) (int, error) {
	return s.CountOPKs(ctx, userID)
}
func (s *PostgresKeyStore) InsertOPKs(ctx context.Context, userID string, opks []OneTimePreKeyUpload) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO x3dh_one_time_prekeys (user_id, key_id, public_key, is_used, created_at)
		VALUES ($1, $2, $3, FALSE, NOW())
		ON CONFLICT (user_id, key_id) DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, opk := range opks {
		if opk.KeyID == "" {
			return fmt.Errorf("opk key_id is required")
		}
		if _, err := stmt.ExecContext(ctx, userID, opk.KeyID, opk.PublicKey); err != nil {
			return fmt.Errorf("insert OPK failed: %w", err)
		}
	}

	return tx.Commit()
}

func (s *PostgresKeyStore) ListUsersWithOPKs(ctx context.Context) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT DISTINCT user_id FROM x3dh_one_time_prekeys
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

func (s *PostgresKeyStore) RotateSignedPreKey(ctx context.Context, userID, newSPK, signature string, expiresAt *time.Time) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE x3dh_signed_prekeys
		SET public_key = $1,
		    signature = $2,
		    created_at = CURRENT_TIMESTAMP,
		    expires_at = $3
		WHERE user_id = $4
	`, newSPK, signature, expiresAt, userID)
	return err
}
