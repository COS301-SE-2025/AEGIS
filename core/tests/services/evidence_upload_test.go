package services

import (
	"os"
	"testing"
	"aegis-api/services/evidence"
	"aegis-api/db"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path/filepath"
)

func init() {
	if err := db.Connect(); err != nil {
		panic("❌ Failed to connect to DB: " + err.Error())
	}
	   os.Setenv("MONGO_URI", "mongodb://admin:mongo_secure_password123@localhost:27017/app_database?authSource=admin")

    if err := db.ConnectMongo(); err != nil {
        panic("❌ MongoDB connection failed: " + err.Error())
    }
	evidence.InitIPFSClient()
}

func TestUploadEvidenceService(t *testing.T) {
	// 1. Use a real test file
	testFile := "/mnt/c/Users/ofent/OneDrive - University of Pretoria/Documents/COS 301/AEGIS - project/AEGIS/core/tests/services/evidence_upload_file.md"


	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("❌ Failed to open test file: %v", err)
	}
	defer file.Close()

	stat, _ := file.Stat()

	// 2. Upload to IPFS
	cid, err := evidence.UploadToIPFS(testFile)
	if err != nil {
		t.Fatalf("❌ IPFS upload failed: %v", err)
	}

	// 3. Construct request
	// 3. Compute checksum
hash := sha256.New()
if _, err := file.Seek(0, 0); err != nil {
	t.Fatalf("❌ Failed to seek file: %v", err)
}
if _, err := io.Copy(hash, file); err != nil {
	t.Fatalf("❌ Failed to compute hash: %v", err)
}
	checksum := hex.EncodeToString(hash.Sum(nil))

	// 4. Construct upload request
	req := evidence.UploadEvidenceRequest{
		CaseID:     "3c336456-aa9a-4298-8b24-cadaa5e41bf2",
		UploadedBy: "8fb89568-3c52-4535-af33-d2f1266def52",
		Filename:   stat.Name(),
		FileType:   "text/markdown",
		IpfsCID:    cid,
		FileSize:   stat.Size(),
		Checksum:   checksum,
		Metadata: map[string]interface{}{
			"source":    "unit test",
			"extension": filepath.Ext(stat.Name()),
		},
		Tags: []string{"test", "markdown"},
	}


	// 4. Call service
	service := evidence.NewEvidenceService()
	result, err := service.UploadEvidence(req)
	if err != nil {
		t.Fatalf("❌ UploadEvidence service failed: %v", err)
	}

	t.Logf("✅ Evidence stored: %s (CID: %s)", result.Filename, result.IpfsCID)
}
