package unit_tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"aegis-api/internal/x3dh"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *x3dh.PostgresKeyStore) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	return db, mock, x3dh.NewPostgresKeyStore(db)
}

// ─── GetIdentityKey ─────────────────────────────────────────────

func TestPostgresKeyStore_GetIdentityKey_Success(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"public_key"}).AddRow("ik_pub")
	mock.ExpectQuery(`SELECT public_key FROM x3dh_identity_keys`).
		WithArgs("user1").
		WillReturnRows(rows)

	ik, err := store.GetIdentityKey(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Equal(t, "ik_pub", ik.PublicKey)
	assert.Equal(t, "user1", ik.UserID)
}

func TestPostgresKeyStore_GetIdentityKey_Error(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	mock.ExpectQuery(`SELECT public_key FROM x3dh_identity_keys`).
		WithArgs("user1").
		WillReturnError(sql.ErrNoRows)

	ik, err := store.GetIdentityKey(context.Background(), "user1")
	assert.Error(t, err)
	assert.Nil(t, ik)
}

// ─── GetSignedPreKey ───────────────────────────────────────────

func TestPostgresKeyStore_GetSignedPreKey_Success(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"public_key", "signature"}).AddRow("spk_pub", "spk_sig")
	mock.ExpectQuery(`SELECT public_key, signature FROM x3dh_signed_prekeys`).
		WithArgs("user1").
		WillReturnRows(rows)

	spk, err := store.GetSignedPreKey(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Equal(t, "spk_pub", spk.PublicKey)
	assert.Equal(t, "spk_sig", spk.Signature)
	assert.Equal(t, "user1", spk.UserID)
}

func TestPostgresKeyStore_GetSignedPreKey_Error(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	mock.ExpectQuery(`SELECT public_key, signature FROM x3dh_signed_prekeys`).
		WithArgs("user1").
		WillReturnError(sql.ErrNoRows)

	spk, err := store.GetSignedPreKey(context.Background(), "user1")
	assert.Error(t, err)
	assert.Nil(t, spk)
}

// ─── ConsumeOneTimePreKey ──────────────────────────────────────

func TestPostgresKeyStore_ConsumeOneTimePreKey_Success(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"id", "key_id", "public_key"}).
		AddRow(1, "key123", "opk_pub")
	mock.ExpectQuery(`SELECT id, key_id, public_key FROM x3dh_one_time_prekeys`).
		WithArgs("user1").
		WillReturnRows(rows)
	mock.ExpectExec(`UPDATE x3dh_one_time_prekeys SET is_used = TRUE`).
		WithArgs("key123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	opk, err := store.ConsumeOneTimePreKey(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Equal(t, "opk_pub", opk.PublicKey)
	assert.Equal(t, "user1", opk.UserID)
}

func TestPostgresKeyStore_ConsumeOneTimePreKey_NoRows(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, key_id, public_key FROM x3dh_one_time_prekeys`).
		WithArgs("user1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	opk, err := store.ConsumeOneTimePreKey(context.Background(), "user1")
	assert.ErrorIs(t, err, x3dh.ErrNoOPKsAvailable)
	assert.Nil(t, opk)
}

// ─── StoreBundle ──────────────────────────────────────────────

func TestPostgresKeyStore_StoreBundle_Success(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO x3dh_identity_keys`).
		WithArgs("user1", "ik_pub").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO x3dh_signed_prekeys`).
		WithArgs("user1", "spk_pub", "spk_sig").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO x3dh_one_time_prekeys`).
		WithArgs("user1", "opk1", "opk_pub1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := x3dh.RegisterBundleRequest{
		UserID:       "user1",
		IdentityKey:  "ik_pub",
		SignedPreKey: "spk_pub",
		SPKSignature: "spk_sig",
		OneTimePreKeys: []x3dh.OneTimePreKeyUpload{
			{KeyID: "opk1", PublicKey: "opk_pub1"},
		},
	}
	err := store.StoreBundle(context.Background(), req, nil)
	assert.NoError(t, err)
}

// ─── CountOPKs / CountAvailableOPKs ───────────────────────────

func TestPostgresKeyStore_CountOPKs(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM x3dh_one_time_prekeys`).
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := store.CountOPKs(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestPostgresKeyStore_CountAvailableOPKs_Delegates(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM x3dh_one_time_prekeys`).
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	count, err := store.CountAvailableOPKs(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Equal(t, 7, count)
}

// ─── InsertOPKs ───────────────────────────────────────────────

func TestPostgresKeyStore_InsertOPKs_Success(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectPrepare(`INSERT INTO x3dh_one_time_prekeys`)
	mock.ExpectExec(`INSERT INTO x3dh_one_time_prekeys`).
		WithArgs("user1", "opk1", "pub1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO x3dh_one_time_prekeys`).
		WithArgs("user1", "opk2", "pub2").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	opks := []x3dh.OneTimePreKeyUpload{
		{KeyID: "opk1", PublicKey: "pub1"},
		{KeyID: "opk2", PublicKey: "pub2"},
	}
	err := store.InsertOPKs(context.Background(), "user1", opks)
	assert.NoError(t, err)
}

// ─── ListUsersWithOPKs ────────────────────────────────────────

func TestPostgresKeyStore_ListUsersWithOPKs(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id"}).AddRow("alice").AddRow("bob")
	mock.ExpectQuery(`SELECT DISTINCT user_id FROM x3dh_one_time_prekeys`).
		WillReturnRows(rows)

	users, err := store.ListUsersWithOPKs(context.Background())
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"alice", "bob"}, users)
}

// ─── RotateSignedPreKey ───────────────────────────────────────

func TestPostgresKeyStore_RotateSignedPreKey(t *testing.T) {
	db, mock, store := setupDB(t)
	defer db.Close()

	expires := time.Now()
	mock.ExpectExec(`UPDATE x3dh_signed_prekeys`).
		WithArgs("new_pub", "new_sig", &expires, "user1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.RotateSignedPreKey(context.Background(), "user1", "new_pub", "new_sig", &expires)
	assert.NoError(t, err)
}
