package evidence_viewer

import (
    "fmt"
    "gorm.io/gorm"
)

type PostgresEvidenceRepository struct {
    DB         *gorm.DB
    IPFSClient *IPFSClient
}

func NewPostgresEvidenceRepository(db *gorm.DB, ipfsClient *IPFSClient) *PostgresEvidenceRepository {
    return &PostgresEvidenceRepository{
        DB:         db,
        IPFSClient: ipfsClient,
    }
}

func (repo *PostgresEvidenceRepository) GetEvidenceFileByID(evidenceID string) (*EvidenceFile, error) {
    var cid string
    result := repo.DB.Model(&EvidenceDTO{}).
        Select("ipfs_cid").
        Where("id = ?", evidenceID).
        Scan(&cid)

    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, result.Error
    }

    if cid == "" {
        return nil, fmt.Errorf("evidence not found")
    }

    content, err := repo.IPFSClient.getEvidence(cid)
    if err != nil {
        return nil, fmt.Errorf("failed to get file from IPFS: %w", err)
    }

    return &EvidenceFile{
        ID:   evidenceID,
        Data: content,
    }, nil
}


func (repo *PostgresEvidenceRepository) GetEvidenceFilesByCaseID(caseID string) ([]EvidenceFile, error) {
    var pairs []EvidenceCIDPair
    result := repo.DB.Model(&EvidenceDTO{}).
        Select("id, ipfs_cid").
        Where("case_id = ?", caseID).
        Scan(&pairs)

    if result.Error != nil {
        return nil, result.Error
    }

    var files []EvidenceFile
    for _, pair := range pairs {
        content, err := repo.IPFSClient.getEvidence(pair.IPFSCID)
        if err != nil {
            return nil, fmt.Errorf("failed to get file for evidence ID %s: %w", pair.ID, err)
        }

        files = append(files, EvidenceFile{
            ID:   pair.ID,
            Data: content,
        })
    }

    return files, nil
}


func (repo *PostgresEvidenceRepository) SearchEvidenceFiles(query string) ([]EvidenceFile, error) {
    var pairs []EvidenceCIDPair

    pattern := "%" + query + "%"
    result := repo.DB.Model(&EvidenceDTO{}).
        Select("id, ipfs_cid").
        Where(
            "filename ILIKE ? OR file_type ILIKE ? OR metadata::text ILIKE ?",
            pattern, pattern, pattern,
        ).
        Scan(&pairs)

    if result.Error != nil {
        return nil, result.Error
    }

    var files []EvidenceFile
    for _, pair := range pairs {
        content, err := repo.IPFSClient.getEvidence(pair.IPFSCID)
        if err != nil {
            return nil, fmt.Errorf("failed to get file for evidence ID %s: %w", pair.ID, err)
        }

        files = append(files, EvidenceFile{
            ID:   pair.ID,
            Data: content,
        })
    }

    return files, nil
}


type EvidenceCIDPair struct {
    ID      string `json:"id"`
    IPFSCID string `json:"ipfs_cid"`
}

func (repo *PostgresEvidenceRepository) GetFilteredEvidenceFiles(
    caseID string,
    filters map[string]interface{},
    sortField, sortOrder string,
) ([]EvidenceFile, error) {
    var pairs []EvidenceCIDPair

    tx := repo.DB.Model(&EvidenceDTO{}).
        Select("id, ipfs_cid").
        Where("case_id = ?", caseID)

    for k, v := range filters {
        tx = tx.Where(fmt.Sprintf("%s = ?", k), v)
    }

    if sortField != "" && (sortOrder == "asc" || sortOrder == "desc") {
        tx = tx.Order(fmt.Sprintf("%s %s", sortField, sortOrder))
    }

    result := tx.Scan(&pairs)
    if result.Error != nil {
        return nil, result.Error
    }

    var files []EvidenceFile
    for _, pair := range pairs {
        content, err := repo.IPFSClient.getEvidence(pair.IPFSCID)
        if err != nil {
            return nil, fmt.Errorf("failed to get file for evidence ID %s: %w", pair.ID, err)
        }

        files = append(files, EvidenceFile{
            ID:   pair.ID,
            Data: content,
        })
    }

    return files, nil
}
