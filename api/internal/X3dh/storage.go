package x3dh

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type KeyStore interface {
	GetIdentityKey(ctx context.Context, userID string) (*IdentityKey, error)
	GetSignedPreKey(ctx context.Context, userID string) (*SignedPreKey, error)
	ConsumeOneTimePreKey(ctx context.Context, userID string) (*OneTimePreKey, error)
	CountOPKs(ctx context.Context, userID string) (int, error)
	StoreBundle(ctx context.Context, req RegisterBundleRequest) error
	InsertOPKs(ctx context.Context, userID string, opks []OneTimePreKeyUpload) error
	CountAvailableOPKs(ctx context.Context, userID string) (int, error)
	ListUsersWithOPKs(ctx context.Context) ([]string, error)
	RotateSignedPreKey(ctx context.Context, userID string, newSPK, signature string, expiresAt *time.Time) error
}

var ErrNoOPKsAvailable = errors.New("no available one-time prekeys")

type PostgresKeyStore struct {
	DB *sql.DB
}

func NewPostgresKeyStore(db *sql.DB) *PostgresKeyStore {
	return &PostgresKeyStore{DB: db}
}

func (s *PostgresKeyStore) GetIdentityKey(ctx context.Context, userID string) (*IdentityKey, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT public_key, private_key
		FROM x3dh_identity_keys
		WHERE user_id = $1
	`, userID)

	var ik IdentityKey
	ik.UserID = userID
	if err := row.Scan(&ik.PublicKey, &ik.PrivateKey); err != nil {
		return nil, fmt.Errorf("failed to fetch identity key: %w", err)
	}
	return &ik, nil
}

func (s *PostgresKeyStore) GetSignedPreKey(ctx context.Context, userID string) (*SignedPreKey, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT public_key, private_key, signature
		FROM x3dh_signed_prekeys
		WHERE user_id = $1
	`, userID)

	var spk SignedPreKey
	spk.UserID = userID
	if err := row.Scan(&spk.PublicKey, &spk.PrivateKey, &spk.Signature); err != nil {
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
		SELECT id, public_key, private_key
		FROM x3dh_one_time_prekeys
		WHERE user_id = $1 AND is_used = FALSE
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE
	`, userID)

	var opk OneTimePreKey
	opk.UserID = userID
	if err := row.Scan(&opk.ID, &opk.PublicKey, &opk.PrivateKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoOPKsAvailable
		}
		return nil, fmt.Errorf("failed to fetch one-time prekey: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE x3dh_one_time_prekeys
		SET is_used = TRUE
		WHERE id = $1
	`, opk.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark one-time prekey as used: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return &opk, nil
}

func (s *PostgresKeyStore) StoreBundle(ctx context.Context, req RegisterBundleRequest, crypto CryptoService) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Encrypt private keys before storing (server-side encryption for now)
	encryptedIKPriv, _ := crypto.Encrypt(req.IdentityKey)   // Normally this should be the actual private key
	encryptedSPKPriv, _ := crypto.Encrypt(req.SignedPreKey) // Same for SPK
	// If real private keys are available, use those instead of public fields

	_, err = tx.ExecContext(ctx, `
		INSERT INTO x3dh_identity_keys (user_id, public_key, private_key)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET public_key = EXCLUDED.public_key, private_key = EXCLUDED.private_key
	`, req.UserID, req.IdentityKey, encryptedIKPriv)
	if err != nil {
		return fmt.Errorf("failed to insert IK: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
	INSERT INTO x3dh_signed_prekeys (user_id, public_key, private_key, signature, created_at)
	VALUES ($1, $2, $3, $4, NOW())
	ON CONFLICT (user_id) DO UPDATE 
	SET public_key = EXCLUDED.public_key,
		signature = EXCLUDED.signature,
		created_at = EXCLUDED.created_at
`, req.UserID, req.SignedPreKey, encryptedSPKPriv, req.SPKSignature)

	if err != nil {
		return fmt.Errorf("failed to insert SPK: %w", err)
	}

	for _, opk := range req.OneTimePreKeys {
		encryptedOPKPriv, _ := crypto.Encrypt(opk.PublicKey) // Same note as above
		_, err := tx.ExecContext(ctx, `
			INSERT INTO x3dh_one_time_prekeys (user_id, public_key, private_key)
			VALUES ($1, $2, $3)
		`, req.UserID, opk.PublicKey, encryptedOPKPriv)
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

func (s *BundleService) CountAvailableOPKs(ctx context.Context, userID string) (int, error) {
	return s.store.CountOPKs(ctx, userID)
}

func (s *PostgresKeyStore) InsertOPKs(ctx context.Context, userID string, opks []OneTimePreKeyUpload) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, opk := range opks {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO x3dh_one_time_prekeys (user_id, public_key)
		VALUES ($1, $2)
		ON CONFLICT (user_id, public_key) DO NOTHING
	`, userID, opk.PublicKey)

		if err != nil {
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
