// ipfs_uploader.go
package evidence

import (
	"bytes"
	"io"

	shell "github.com/ipfs/go-ipfs-api"
)

var ipfs *shell.Shell

func InitIPFSClient() {
	ipfs = shell.NewShell("localhost:5001")
}

type IPFSShellClient struct {
	shell *shell.Shell
}

func NewIPFSShellClient(addr string) *IPFSShellClient {
	return &IPFSShellClient{
		shell: shell.NewShell(addr),
	}
}
func (c *IPFSShellClient) Upload(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	cid, err := c.shell.Add(reader)
	if err != nil {
		return "", err
	}
	return cid, nil
}

func (c *IPFSShellClient) Download(cid string) ([]byte, error) {
	reader, err := c.shell.Cat(cid)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
