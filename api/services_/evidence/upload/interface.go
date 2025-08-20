package upload

import "io"

// IPFSClientImp defines the contract for uploading/downloading files via IPFS.
type IPFSClientImp interface {
	UploadFile(file io.Reader) (string, error)
	Download(cid string) (io.ReadCloser, error)
}
