package upload

import "io"

// IPFSClientImp defines the contract for uploading/downloading files via IPFS.
type IPFSClientImp interface {
	UploadFile(path string) (string, error)
	Download(cid string) (io.ReadCloser, error)
}
