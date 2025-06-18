package unit_tests

import (
    "testing"
    "time"

    "aegis-api/services/Evidence_Viewer"
    "aegis-api/models"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
)

func TestGetEvidenceByCase(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()

    repo := &Evidence_Viewer.PostgresEvidenceRepository{DB: db}

    rows := sqlmock.NewRows([]string{
        "id", "case_id", "uploaded_by", "filename", "file_type",
        "ipfs_cid", "file_size", "checksum", "metadata", "uploaded_at",
    }).AddRow(
        "1", "case123", "officer1", "photo.jpg", "image",
        "cid123", 2048, "abc123", "{}", time.Now(),
    )

    mock.ExpectQuery(`SELECT id, case_id, uploaded_by, filename, file_type, ipfs_cid, file_size, checksum, metadata, uploaded_at FROM evidence WHERE case_id = \$1`).
        WithArgs("case123").
        WillReturnRows(rows)

    evidences, err := repo.GetEvidenceByCase("case123")
    assert.NoError(t, err)
    assert.Len(t, evidences, 1)
    assert.Equal(t, "photo.jpg", evidences[0].Filename)
}

func TestGetEvidenceByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &Evidence_Viewer.PostgresEvidenceRepository{DB: db}

	expected := models.EvidenceDTO{
		ID:         "ev123",
		CaseID:     "case456",
		UploadedBy: "officer1",
		Filename:   "file1.jpg",
		FileType:   "image",
		IPFSCID:    "cid456",
		FileSize:   2048,
		Checksum:   "checksum",
		Metadata:   "{}",
		UploadedAt: time.Now().Format(time.RFC3339),
	}

	mock.ExpectQuery(`SELECT id, case_id, uploaded_by, filename, file_type, ipfs_cid, file_size, checksum, metadata, uploaded_at FROM evidence WHERE id = \$1`).
		WithArgs("ev123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "case_id", "uploaded_by", "filename", "file_type", "ipfs_cid",
			"file_size", "checksum", "metadata", "uploaded_at",
		}).AddRow(
			expected.ID, expected.CaseID, expected.UploadedBy, expected.Filename,
			expected.FileType, expected.IPFSCID, expected.FileSize,
			expected.Checksum, expected.Metadata, expected.UploadedAt,
		))

	result, err := repo.GetEvidenceByID("ev123")
	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Filename, result.Filename)
}

func TestSearchEvidence(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &Evidence_Viewer.PostgresEvidenceRepository{DB: db}

	mockRows := sqlmock.NewRows([]string{"id", "filename", "file_type", "ipfs_cid"}).
		AddRow("1", "bodycam.jpg", "image", "cid1").
		AddRow("2", "transcript.txt", "text", "cid2")

	mock.ExpectQuery(`SELECT id, filename, file_type, ipfs_cid FROM evidence WHERE filename ILIKE \$1 OR file_type ILIKE \$1 OR metadata::text ILIKE \$1`).
		WithArgs("%body%").
		WillReturnRows(mockRows)

	results, err := repo.SearchEvidence("body")
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "bodycam.jpg", results[0].Filename)
	assert.Equal(t, "cid2", results[1].IPFSCID)
}

func TestGetFilteredEvidence(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &Evidence_Viewer.PostgresEvidenceRepository{DB: db}

	filters := map[string]interface{}{
		"file_type": "image",
	}

	mockRows := sqlmock.NewRows([]string{"id", "filename", "file_type", "ipfs_cid"}).
		AddRow("1", "evidence.jpg", "image", "cid-image")

	mock.ExpectQuery(`SELECT id, filename, file_type, ipfs_cid FROM evidence WHERE case_id = \$1 AND file_type = \$2 ORDER BY uploaded_at desc`).
		WithArgs("case789", "image").
		WillReturnRows(mockRows)

	results, err := repo.GetFilteredEvidence("case789", filters, "uploaded_at", "desc")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "evidence.jpg", results[0].Filename)
}


