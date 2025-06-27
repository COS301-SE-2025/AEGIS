package evidence_download

import (
	"aegis-api/services_/evidence/metadata"
	upload "aegis-api/services_/evidence/upload"
)

type Service struct {
	Repo metadata.Repository
	IPFS upload.IPFSClientImp
}
