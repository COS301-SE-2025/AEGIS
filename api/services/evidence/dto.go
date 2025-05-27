package evidence
type UploadEvidenceRequest struct {
	CaseID     string                 `json:"case_id"`
	UploadedBy string                 `json:"uploaded_by"`
	Filename   string                 `json:"filename"`
	FileType   string                 `json:"file_type"`
	IpfsCID    string                 `json:"ipfs_cid"`
	FileSize   int64                  `json:"file_size"`
	Checksum   string                 `json:"checksum"`
	Metadata   map[string]interface{} `json:"metadata"`
	Tags       []string               `json:"tags"`
}
