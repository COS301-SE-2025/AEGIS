package integration

import (
	"os"
	"testing"

	"aegis-api/services_/evidence/upload"

	"github.com/stretchr/testify/require"
)

func TestUploadFileIntegration(t *testing.T) {
	// Step 1: Create a temporary file with test content
	tmpFile, err := os.CreateTemp("", "upload-test-*.md")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // clean up temp file after test

	content := []byte("# Test File for IPFS Upload\nHello, AEGIS!")
	_, err = tmpFile.Write(content)
	require.NoError(t, err)
	tmpFile.Close()

	// Step 2: Open file again as io.Reader
	fileReader, err := os.Open(tmpFile.Name())
	require.NoError(t, err)
	defer fileReader.Close()

	// Step 3: Set up real IPFS client and service
	ipfsClient := upload.NewIPFSClient("http://localhost:5001") // adjust if your IPFS runs elsewhere
	service := upload.NewEvidenceService(ipfsClient)

	// Step 4: Upload file to IPFS
	cid, err := service.UploadFile(fileReader)
	require.NoError(t, err)
	require.NotEmpty(t, cid)

	t.Logf("Successfully uploaded file to IPFS. CID: %s", cid)
}
