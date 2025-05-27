// tests/services/integration_tests/upload_evidence_test.go
package integration_tests

import (
    "testing"
    //"time"

    "github.com/google/uuid"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"

    "aegis-api/services/evidence"
)

// fakeIPFSClient satisfies IPFSClient but does nothing.
type fakeIPFSClient struct{}

func (f *fakeIPFSClient) Upload(path string) (string, error)   { return "fakeCID", nil }
func (f *fakeIPFSClient) Download(cid string) ([]byte, error) { return []byte(""), nil }

// fakeLogger satisfies EvidenceLogger but does no real logging.
type fakeLogger struct{}

func (l *fakeLogger) Log(userID, evidenceID, filename string) error { return nil }

// dummyRepo implements only the bits of EvidenceRepository needed for UploadEvidence; other methods are stubs.
type dummyRepo struct{ db *gorm.DB }

func (r *dummyRepo) SaveEvidence(e *evidence.Evidence) error {
    return r.db.Create(e).Error
}
func (r *dummyRepo) AttachTags(e *evidence.Evidence, tags []string) error {
    return nil
}
func (r *dummyRepo) FindByID(uuid.UUID) (*evidence.Evidence, error)    { return nil, nil }
func (r *dummyRepo) DeleteByID(uuid.UUID) error                          { return nil }
func (r *dummyRepo) FindByCase(uuid.UUID) ([]evidence.Evidence, error)   { return nil, nil }
func (r *dummyRepo) FindByUser(uuid.UUID) ([]evidence.Evidence, error)   { return nil, nil }
func (r *dummyRepo) PreloadMetadata(uuid.UUID) (*evidence.Evidence, error) { return nil, nil }

func TestUploadEvidence_Integration(t *testing.T) {
    // 1) In-memory SQLite via GORM
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open sqlite: %v", err)
    }

    // 2) Create a SQLite-compatible evidence table manually
    ddl := `
    CREATE TABLE IF NOT EXISTS evidence (
      id TEXT PRIMARY KEY,
      case_id TEXT NOT NULL,
      uploaded_by TEXT NOT NULL,
      filename TEXT NOT NULL,
      file_type TEXT NOT NULL,
      ipfs_cid TEXT NOT NULL,
      file_size INTEGER CHECK(file_size >= 0),
      checksum TEXT NOT NULL,
      metadata TEXT,
      uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    `
    if err := db.Exec(ddl).Error; err != nil {
        t.Fatalf("manual migrate failed: %v", err)
    }


    // 3) Build service
    repo := &dummyRepo{db: db}
    ipfs := &fakeIPFSClient{}
    logger := &fakeLogger{}
    svc := evidence.NewEvidenceService(ipfs, repo, logger)

    // 4) Prepare a valid request (no Ctx field)
    req := evidence.UploadEvidenceRequest{
        CaseID:     uuid.New().String(),
        UploadedBy: uuid.New().String(),
        Filename:   "testfile.txt",
        FileType:   "text/plain",
        IpfsCID:    "fakeCID",
        FileSize:   42,
        Checksum:   "abc123",
        Metadata:   map[string]interface{}{"foo": "bar"},
        Tags:       []string{"alpha", "beta"},
    }

    // 5) Call UploadEvidence
    created, err := svc.UploadEvidence(req)
    if err != nil {
        t.Fatalf("UploadEvidence returned error: %v", err)
    }

    // 6) Sanity checks on returned object
    if created.Filename != req.Filename {
        t.Errorf("Filename = %q; want %q", created.Filename, req.Filename)
    }
    if created.IpfsCID != req.IpfsCID {
        t.Errorf("IpfsCID = %q; want %q", created.IpfsCID, req.IpfsCID)
    }
    if created.FileSize != req.FileSize {
        t.Errorf("FileSize = %d; want %d", created.FileSize, req.FileSize)
    }

    // 7) Read it back from the DB
    var persisted evidence.Evidence
    if err := db.First(&persisted, "id = ?", created.ID).Error; err != nil {
        t.Fatalf("could not load evidence from DB: %v", err)
    }

    // 8) Assert the DB row matches
    if persisted.Filename != req.Filename {
        t.Errorf("DB.Filename = %q; want %q", persisted.Filename, req.Filename)
    }
    if persisted.IpfsCID != req.IpfsCID {
        t.Errorf("DB.IpfsCID = %q; want %q", persisted.IpfsCID, req.IpfsCID)
    }
    if persisted.FileSize != req.FileSize {
        t.Errorf("DB.FileSize = %d; want %d", persisted.FileSize, req.FileSize)
    }
}
