package upload

import (
	"io"

	shell "github.com/ipfs/go-ipfs-api"
)

// ipfsClientImpl is a concrete implementation of IPFSClientImp.
type ipfsClientImpl struct {
	shell *shell.Shell
}

// NewIPFSClient creates and returns a new IPFS client.
func NewIPFSClient(api string) IPFSClientImp {
	if api == "" {
		api = "http://ipfs:5001"
	}
	return &ipfsClientImpl{
		shell: shell.NewShell(api),
	}
}

func (c *ipfsClientImpl) UploadFile(file io.Reader) (string, error) {
	return c.shell.Add(file)
}

func (c *ipfsClientImpl) Download(cid string) (io.ReadCloser, error) {
	return c.shell.Cat(cid)
}
