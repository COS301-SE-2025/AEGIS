// ipfs_uploader.go
package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

// ipfsUploader implements the IPFSUploader interface
type ipfsUploader struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewIPFSUploader creates a new IPFS uploader instance
func NewIPFSUploader(baseURL, apiKey string) IPFSUploader {
	return &ipfsUploader{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UploadFile uploads a multipart file to IPFS
func (u *ipfsUploader) UploadFile(ctx context.Context, file multipart.File, fileName string) (*IPFSUploadResult, error) {
	// Reset file pointer to beginning
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// Read file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return u.UploadBytes(ctx, data, fileName)
}

// UploadBytes uploads byte data to IPFS
func (u *ipfsUploader) UploadBytes(ctx context.Context, data []byte, fileName string) (*IPFSUploadResult, error) {
	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fileWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = fileWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	// Add additional parameters if needed
	writer.WriteField("pin", "true") // Pin the file by default

	writer.Close()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", u.baseURL+"/api/v0/add", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if u.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+u.apiKey)
	}

	// Send request
	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("IPFS upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ipfsResp struct {
		Hash string `json:"Hash"`
		Name string `json:"Name"`
		Size string `json:"Size"`
	}

	sizeInt, err := strconv.ParseInt(ipfsResp.Size, 10, 64)
	if err != nil {
		sizeInt = int64(len(data)) // fallback to actual file data size
	}

	if err := json.NewDecoder(resp.Body).Decode(&ipfsResp); err != nil {
		return nil, fmt.Errorf("failed to parse IPFS response: %w", err)
	}

	result := &IPFSUploadResult{
		Hash:     ipfsResp.Hash,
		URL:      u.GetFileURL(ipfsResp.Hash),
		Size:     sizeInt,
		FileName: fileName,
	}

	return result, nil
}

// GetFileURL generates a URL for accessing an IPFS file
func (u *ipfsUploader) GetFileURL(hash string) string {

	return fmt.Sprintf("http://localhost:8081/ipfs/%s", hash)
}

// DeleteFile unpins a file from IPFS (equivalent to deletion)
func (u *ipfsUploader) DeleteFile(ctx context.Context, hash string) error {
	// Note: IPFS doesn't support direct deletion, but you can unpin files
	// which makes them eligible for garbage collection
	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/api/v0/pin/rm?arg=%s", u.baseURL, hash), nil)
	if err != nil {
		return fmt.Errorf("failed to create unpin request: %w", err)
	}

	if u.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+u.apiKey)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unpin file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to unpin file, status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ValidateFileType checks if a file type is allowed
func (u *ipfsUploader) ValidateFileType(fileName string, allowedTypes []string) bool {
	if len(allowedTypes) == 0 {
		return true // No restrictions
	}

	// Extract file extension
	ext := ""
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			ext = fileName[i:]
			break
		}
	}

	for _, allowedType := range allowedTypes {
		if ext == allowedType {
			return true
		}
	}

	return false
}

// ValidateFileSize checks if file size is within limits
func (u *ipfsUploader) ValidateFileSize(size int64, maxSize int64) bool {
	if maxSize <= 0 {
		return true // No size limit
	}
	return size <= maxSize
}
