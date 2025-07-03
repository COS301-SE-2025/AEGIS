package evidence_viewer

type EvidenceResponse struct {
    ID         string `json:"id"`
    CaseID     string `json:"case_id"`
    Filename   string `json:"filename"`
    FileType   string `json:"file_type"`
    IPFSCID    string `json:"ipfs_cid"`
    UploadedBy string `json:"uploaded_by"`
    Metadata   string `json:"metadata"`
    UploadedAt string `json:"uploaded_at"`
}