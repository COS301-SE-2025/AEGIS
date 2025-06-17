package evidence

import (
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSClient is a client for interacting with an IPFS node.
// // It provides methods to upload files to IPFS and retrieve their CIDs (Content Identifiers).
type IPFSClient struct {
	shell *shell.Shell
}

// NewIPFSClient creates a new IPFS client instance.
// It takes the API endpoint as an argument and returns a pointer to IPFSClient.
// If no API endpoint is provided, it defaults to "http://ipfs:5001".
func NewIPFSClient(api string) *IPFSClient {
	return &IPFSClient{
		shell: shell.NewShell("http://ipfs:5001"),
	}
}

// // UploadFile uploads a file to IPFS and returns the CID (Content Identifier).
// It takes the file path as an argument and returns the CID or an error.
func (c *IPFSClient) UploadFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	cid, err := c.shell.Add(file)
	if err != nil {
		return "", err
	}
	return cid, nil
}
