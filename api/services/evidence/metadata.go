package evidence

type EvidenceMetadata struct {
	ID         string                 `json:"id"`
	Filename   string                 `json:"filename"`
	FileType   string                 `json:"file_type"`
	IpfsCID    string                 `json:"ipfs_cid"`
	FileSize   int64                  `json:"file_size"`
	Checksum   string                 `json:"checksum"`
	Metadata   map[string]interface{} `json:"metadata"`  // JSONB field
	Tags       []string               `json:"tags"`       // From evidence_tags join
	CaseID     string                 `json:"case_id"`
	UploadedBy string                 `json:"uploaded_by"`
	UploadedAt string                 `json:"uploaded_at"`
}
