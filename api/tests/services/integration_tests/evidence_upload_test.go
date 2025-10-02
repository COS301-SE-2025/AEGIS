package integration_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// helper to insert a minimal case row (FKs are enforced on evidence)
// func insertCaseRow(t *testing.T, id uuid.UUID) {
// 	t.Helper()
// 	err := pgDB.Exec(`
// 		INSERT INTO cases (id, title, team_name, created_by, tenant_id, team_id, created_at)
// 		VALUES (?, ?, ?, ?, ?, ?, NOW())
// 		ON CONFLICT (id) DO NOTHING
// 	`, id, "EvCase "+id.String()[:8], "test-team", FixedUserID, FixedTenantID, FixedTeamID).Error
// 	require.NoError(t, err)
// }

func doMultipartRequest(t *testing.T, method, url string, fields map[string]string, fileField, fileName string, fileContent []byte) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// fields
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}

	// file
	fw, err := w.CreateFormFile(fileField, fileName)
	require.NoError(t, err)
	_, err = fw.Write(fileContent)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// func Test_EvidenceUpload_Success(t *testing.T) {
// 	caseID := uuid.New()
// 	setupCaseForTest(t, caseID)
// 	insertCaseRow(t, caseID)

// 	// file content
// 	content := []byte("hello-evidence-upload")
// 	sum := sha256.Sum256(content)
// 	wantSHA256 := hex.EncodeToString(sum[:])

// 	fields := map[string]string{
// 		"case_id":   caseID.String(),
// 		"filename":  "hello.txt",
// 		"file_type": "text/plain",
// 		"metadata":  `{"source":"integration-test"}`,
// 	}

// 	w := doMultipartRequest(t, "POST", "/evidence", fields, "file", "hello.txt", content)
// 	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

// 	var resp map[string]any
// 	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

// 	// sanity
// 	require.Equal(t, "hello.txt", resp["filename"])
// 	require.Equal(t, fmt.Sprintf("%d", len(content)), fmt.Sprintf("%v", resp["file_size"]))

// 	// check checksum and CID
// 	gotChecksum, _ := resp["checksum"].(string)
// 	require.Equal(t, wantSHA256, gotChecksum)

// 	gotCID, _ := resp["ipfs_cid"].(string)
// 	require.Equal(t, "bafybeifaketestcidxxxxxxxxxxxxxxxxxxxxxxxx", gotCID)

// 	// verify row in DB has the metadata with sha256/md5 keys
// 	type row struct{ Metadata string }
// 	var r row
// 	err := pgDB.Raw(`SELECT metadata FROM evidence WHERE case_id = ? AND filename = ?`, caseID, "hello.txt").Scan(&r).Error
// 	require.NoError(t, err)

// 	var meta map[string]any
// 	require.NoError(t, json.Unmarshal([]byte(r.Metadata), &meta))
// 	require.NotEmpty(t, meta["sha256"])
// 	require.NotEmpty(t, meta["md5"])
// 	require.Equal(t, "integration-test", meta["source"])
// }

func Test_EvidenceUpload_InvalidCaseID(t *testing.T) {
	content := []byte("x")
	fields := map[string]string{
		"case_id": "not-a-uuid",
	}
	w := doMultipartRequest(t, "POST", "/evidence", fields, "file", "x.bin", content)
	require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func Test_EvidenceUpload_MissingFile(t *testing.T) {
	caseID := uuid.New()
	insertCaseRow(t, caseID)

	// no file part
	var buf bytes.Buffer
	mpw := multipart.NewWriter(&buf)
	require.NoError(t, mpw.WriteField("case_id", caseID.String()))
	require.NoError(t, mpw.Close())

	req := httptest.NewRequest("POST", "/evidence", &buf)
	req.Header.Set("Content-Type", mpw.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
}
