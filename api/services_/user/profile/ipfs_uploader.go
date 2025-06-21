package profile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// IPFSUpload handles uploading files to IPFS
type IPFSUpload struct {
	nodeURL string // IPFS node URL ( "http://localhost:5001")
}

// NewIPFSUploader creates a new IPFS uploader instance
func NewIPFSUploader(nodeURL string) *IPFSUpload {
	if nodeURL == "" {
		nodeURL = "http://localhost:5001" // Default IPFS node URL
	}
	return &IPFSUpload{
		nodeURL: nodeURL,
	}
}

// UploadResponse represents the response from IPFS add operation
type UploadResponse struct {
	Hash string `json:"Hash"`
	Name string `json:"Name"`
	Size string `json:"Size"`
}

// UploadFile uploads a file to IPFS and returns the hash
func (u *IPFSUpload) UploadFile(filename string, fileData []byte) (*UploadResponse, error) {
	// Validate file type for profile pictures
	if !u.isValidImageType(filename) {
		return nil, fmt.Errorf("invalid file type: only JPEG, PNG, GIF, and WebP images are allowed")
	}

	// Validate file size (limit to 5MB)
	const maxSize = 5 * 1024 * 1024 // 5MB
	if len(fileData) > maxSize {
		return nil, fmt.Errorf("file size exceeds maximum limit of 5MB")
	}

	// Create multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the file to the form
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = part.Write(fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	// Close the writer to finalize the form
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close form writer: %w", err)
	}

	// Make the HTTP request to IPFS
	url := fmt.Sprintf("%s/api/v0/add", u.nodeURL)  ///
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("IPFS upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var uploadResp UploadResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// IPFS returns newline-delimited JSON, we want the first line
	lines := strings.Split(string(body), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty response from IPFS")
	}

	// Parse the JSON response
	err = json.Unmarshal([]byte(lines[0]), &uploadResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPFS response: %w", err)
	}

	return &uploadResp, nil
}

// GetIPFSURL returns the public IPFS URL for a given hash
func (u *IPFSUpload) GetIPFSURL(hash string) string {
	// Using a public IPFS gateway - you might want to use your own gateway
	return fmt.Sprintf("https://ipfs.io/ipfs/%s", hash)
}

// DeleteFile removes a file from IPFS (note: this only unpins it from your node)
func (u *IPFSUpload) DeleteFile(hash string) error {
	url := fmt.Sprintf("%s/api/v0/pin/rm?arg=%s", u.nodeURL, hash)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create unpin request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unpin from IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("IPFS unpin failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// isValidImageType checks if the file extension is a valid image type
func (u *IPFSUpload) isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	return validTypes[ext]
}

// UploadProfilePicture is a convenience method specifically for profile picture uploads
func (u *IPFSUpload) UploadProfilePicture(filename string, fileData []byte, userID string) (string, error) {
	// Add user ID to filename to make it unique
	ext := filepath.Ext(filename)
	uniqueFilename := fmt.Sprintf("%s_profile%s", userID, ext)
	
	uploadResp, err := u.UploadFile(uniqueFilename, fileData)
	if err != nil {
		return "", err
	}

	// Return the public IPFS URL
	return u.GetIPFSURL(uploadResp.Hash), nil
}