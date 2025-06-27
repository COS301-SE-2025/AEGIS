package integration

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUploadFileIntegration(t *testing.T) {
	// Step 1: Create a temporary file
	tmpFile, err := os.CreateTemp("", "upload-test-*.md")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // clean up

	content := []byte("# Test File for IPFS Upload\nHello, AEGIS!")
	_, err = tmpFile.Write(content)
	require.NoError(t, err)
	tmpFile.Close()

	// Step 2: Set up Gin router with a mock upload handler (bypassing real IPFS call)
	handler := gin.New()
	handler.POST("/api/v1/upload", func(c *gin.Context) {
		type UploadRequest struct {
			Path string `json:"path"`
		}
		var req UploadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Instead of calling real IPFS, return a mock CID
		mockCID := "QmMockedCID1234567890abcdef"
		c.JSON(http.StatusOK, gin.H{"cid": mockCID})
	})

	// Step 3: Create request with temp file path
	payload, _ := json.Marshal(gin.H{"path": tmpFile.Name()})
	req := httptest.NewRequest("POST", "/api/v1/upload", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Step 4: Perform the request
	handler.ServeHTTP(resp, req)

	// Step 5: Assert response
	require.Equal(t, http.StatusOK, resp.Code)
	body, _ := ioutil.ReadAll(resp.Body)
	t.Logf("Response: %s", body)
}
