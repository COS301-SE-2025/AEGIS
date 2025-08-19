package integration_test

// add imports
import (
	"io"
	"net/http"

	evidl "aegis-api/services_/evidence/evidence_download"
	upload "aegis-api/services_/evidence/upload"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// existing test repo stays the same (testMetadataRepo)

// NEW: download endpoint
func registerEvidenceDownloadTestEndpoints(r *gin.Engine) {
	repo := &testMetadataRepo{db: pgDB}
	if testIPFS == nil {
		testIPFS = newFakeIPFS()
	}
	var ipfs upload.IPFSClientImp = testIPFS
	dl := evidl.NewService(repo, ipfs)

	r.GET("/evidence/:id/download", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		filename, reader, filetype, err := dl.DownloadEvidence(id)
		if err != nil {
			// not found or other repo error
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		defer reader.Close()

		if filetype == "" {
			filetype = "application/octet-stream"
		}
		c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
		c.Header("Content-Type", filetype)
		c.Status(http.StatusOK)
		_, _ = io.Copy(c.Writer, reader)
	})
}
