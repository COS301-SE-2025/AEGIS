package evidence

import (
	"bytes"
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"io"
	"os"
)

var ipfs *shell.Shell

func InitIPFSClient() {
	ipfs = shell.NewShell("localhost:5001")
}

func UploadToIPFS(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file into memory
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Upload to IPFS
	cid, err := ipfs.Add(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", fmt.Errorf("IPFS upload failed: %w", err)
	}

	return cid, nil
}
