// --- imports

package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	metadata "aegis-api/services_/evidence/metadata"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---- fake IPFS client for tests ----
//type fakeIPFS struct{}

func newFakeIPFS() *fakeIPFS { return &fakeIPFS{store: map[string][]byte{}} }

// Shared fake IPFS with in-memory store
type fakeIPFS struct {
	mu    sync.Mutex
	store map[string][]byte
}

// single shared instance for all endpoints
var testIPFS *fakeIPFS

func (f *fakeIPFS) UploadFile(r io.Reader) (string, error) {
	// consume full stream to mimic real upload
	_, _ = io.Copy(io.Discard, r)
	// deterministic fake CID
	return "bafybeifaketestcidxxxxxxxxxxxxxxxxxxxxxxxx", nil
}

// Implement the missing Download method.
// Most IPFS client interfaces use (string) -> (io.ReadCloser, error).
// If your interface returns io.Reader instead, just change the return type accordingly.
func (f *fakeIPFS) Download(cid string) (io.ReadCloser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if b, ok := f.store[cid]; ok {
		return io.NopCloser(bytes.NewReader(b)), nil
	}
	// if not found, still return something
	return io.NopCloser(bytes.NewReader([]byte("fake-content"))), nil
}

// ---- gorm-backed repo implementing metadata.Repository ----
type testMetadataRepo struct{ db *gorm.DB }

func (r *testMetadataRepo) SaveEvidence(e *metadata.Evidence) error {
	// metadata.Evidence.Metadata is a JSON string; cast to jsonb on insert
	return r.db.Exec(`
		INSERT INTO evidence
		  (id, case_id, uploaded_by, filename, file_type, ipfs_cid, file_size, checksum, metadata, tenant_id, team_id)
		VALUES
		  (?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb, ?, ?)
	`, e.ID, e.CaseID, e.UploadedBy, e.Filename, e.FileType, e.IpfsCID, e.FileSize, e.Checksum, e.Metadata, e.TenantID, e.TeamID).Error
}

func (r *testMetadataRepo) FindEvidenceByCaseID(caseID uuid.UUID) ([]metadata.Evidence, error) {
	var rows []struct {
		ID         uuid.UUID
		Filename   string
		FileType   string
		IpfsCID    string
		FileSize   int64
		Checksum   string
		Metadata   string
		UploadedBy uuid.UUID
		TenantID   uuid.UUID
		TeamID     uuid.UUID
	}
	if err := r.db.Raw(`
		SELECT id, filename, file_type, ipfs_cid, file_size, checksum, metadata, uploaded_by, tenant_id, team_id
		FROM evidence WHERE case_id = ?
	`, caseID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]metadata.Evidence, 0, len(rows))
	for _, x := range rows {
		out = append(out, metadata.Evidence{
			ID:         x.ID,
			Filename:   x.Filename,
			FileType:   x.FileType,
			IpfsCID:    x.IpfsCID,
			FileSize:   x.FileSize,
			Checksum:   x.Checksum,
			Metadata:   x.Metadata,
			UploadedBy: x.UploadedBy,
			TenantID:   x.TenantID,
			TeamID:     x.TeamID,
			CaseID:     caseID,
		})
	}
	return out, nil
}

func (r *testMetadataRepo) FindEvidenceByID(id uuid.UUID) (*metadata.Evidence, error) {
	row := r.db.Raw(`
		SELECT case_id, filename, file_type, ipfs_cid, file_size, checksum, metadata, uploaded_by, tenant_id, team_id
		FROM evidence WHERE id = ?
	`, id).Row()
	var ev metadata.Evidence
	ev.ID = id
	if err := row.Scan(&ev.CaseID, &ev.Filename, &ev.FileType, &ev.IpfsCID, &ev.FileSize, &ev.Checksum, &ev.Metadata, &ev.UploadedBy, &ev.TenantID, &ev.TeamID); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &ev, nil
}

// AppendEvidenceLog implements metadata.Repository for testMetadataRepo.
// Adjusted to match expected signature: (*metadata.EvidenceLog) error
func (r *testMetadataRepo) AppendEvidenceLog(logEntry *metadata.EvidenceLog) error {
	// Use proper timestamp handling
	timestamp := logEntry.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	details := logEntry.Details
	if details == "" {
		details = "{}"
	}

	return r.db.Exec(`
        INSERT INTO evidence_log (evidence_id, action, timestamp, details)
        VALUES (?, ?, ?, ?)
    `, logEntry.EvidenceID, logEntry.Action, timestamp, details).Error
}

// ---- register the upload endpoint ----
func registerEvidenceTestEndpoints(r *gin.Engine) {
	repo := &testMetadataRepo{db: pgDB}
	ipfs := &fakeIPFS{}
	svc := metadata.NewService(repo, ipfs)

	// POST /evidence (multipart/form-data):
	// fields: file (required), case_id, filename (opt), file_type (opt), metadata (JSON string, opt)
	r.POST("/evidence", func(c *gin.Context) {
		// auth context
		uidStr, tidStr, gidStr := c.GetString("userID"), c.GetString("tenantID"), c.GetString("teamID")
		uid, err1 := uuid.Parse(uidStr)
		tid, err2 := uuid.Parse(tidStr)
		gid, err3 := uuid.Parse(gidStr)
		if err1 != nil || err2 != nil || err3 != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "auth context missing"})
			return
		}

		// multipart
		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}
		f, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer f.Close()

		caseID, err := uuid.Parse(c.PostForm("case_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case_id"})
			return
		}

		filename := c.PostForm("filename")
		if filename == "" {
			filename = fileHeader.Filename
		}

		fileType := c.PostForm("file_type")
		if fileType == "" {
			fileType = http.DetectContentType(make([]byte, 0))
		}

		metaStr := c.PostForm("metadata")
		meta := map[string]string{}
		if metaStr != "" {
			if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "metadata must be JSON"})
				return
			}
		}

		req := metadata.UploadEvidenceRequest{
			FileData:   f,
			CaseID:     caseID,
			UploadedBy: uid,
			TenantID:   tid,
			TeamID:     gid,
			Filename:   filename,
			FileType:   fileType,
			FileSize:   fileHeader.Size,
			Metadata:   meta,
		}
		if err := svc.UploadEvidence(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// fetch the row we just inserted to include IDs in response
		var id uuid.UUID
		var cid, checksum string
		var size int64
		row := pgDB.Raw(`
			SELECT id, ipfs_cid, checksum, file_size
			FROM evidence WHERE case_id = ? AND filename = ?
			ORDER BY uploaded_at DESC LIMIT 1
		`, caseID, filename).Row()
		if err := row.Scan(&id, &cid, &checksum, &size); err != nil {
			c.JSON(http.StatusCreated, gin.H{"ok": true}) // inserted, but no readback
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":        id.String(),
			"case_id":   caseID.String(),
			"ipfs_cid":  cid,
			"checksum":  checksum,
			"file_size": size,
			"filename":  filename,
			"uploaded":  time.Now().Format(time.RFC3339),
		})
	})
}

// Add this method to your testMetadataRepo to fix the interface implementation
func (r *testMetadataRepo) GetLastEvidenceLog(evidenceID uuid.UUID) (*metadata.EvidenceLog, error) {
	var logEntry struct {
		EvidenceID uuid.UUID
		Action     string
		Timestamp  time.Time
		Details    string
		CreatedAt  time.Time
	}

	err := r.db.Raw(`
		SELECT evidence_id, action, timestamp, details, created_at
		FROM evidence_log 
		WHERE evidence_id = ? 
		ORDER BY created_at DESC 
		LIMIT 1
	`, evidenceID).Scan(&logEntry).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound || err == sql.ErrNoRows {
			return nil, nil // Return nil if no log found
		}
		return nil, err
	}

	return &metadata.EvidenceLog{
		EvidenceID: logEntry.EvidenceID,
		Action:     logEntry.Action,
		Timestamp:  logEntry.Timestamp,
		Details:    logEntry.Details,
	}, nil
}
