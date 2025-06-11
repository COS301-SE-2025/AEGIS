package evidence

import (
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

type IPFSClient struct {
	shell *shell.Shell
}

func NewIPFSClient(api string) *IPFSClient {
	return &IPFSClient{
		shell: shell.NewShell("http://ipfs:5001"),
	}
}

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
