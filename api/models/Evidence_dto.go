package models



type EvidenceDTO struct {
    ID         string `json:"id"`
    CaseID     string `json:"case_id"`
    UploadedBy string `json:"uploaded_by"`
    Filename   string `json:"filename"`
    FileType   string `json:"file_type"`
    IPFSCID    string `json:"ipfs_cid"`
    FileSize   int64  `json:"file_size"`
    Checksum   string `json:"checksum"`
    Metadata   string `json:"metadata"`  
    UploadedAt string `json:"uploaded_at"`
}



