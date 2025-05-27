package Evidence_Viewer

import (
    "io"
    "fmt"
    "github.com/ipfs/go-ipfs-api"
)

//IPFSClient is a struct that holds the IPFS shell client
type IPFSClient struct {
    Shell *shell.Shell // IPFS shell client
}

// NewIPFSClient initializes a new IPFS client connected to the local IPFS node
func NewIPFSClient() *IPFSClient {
	return &IPFSClient{
		Shell: shell.NewShell("localhost:5001"), // Connects to IPFS running in Docker
	}
}

func (client *IPFSClient) getEvidence(cid string ) ([]byte, error) {
    // Fetch the file from IPFS using the provided CID
    file, err := client.Shell.Cat(cid)
    if err != nil {
        return nil, fmt.Errorf("failed to get file from IPFS: %w", err)
    }
    defer file.Close()

    // Read the content of the file
    content, err := io.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("failed to read file content: %w", err)
    }

    return content, nil
}

