package evidence_viewer

import (
    "gorm.io/gorm"
    "fmt"
)

type PostgresEvidenceRepository struct {
    DB *gorm.DB
}

func NewPostgresEvidenceRepository(db *gorm.DB) *PostgresEvidenceRepository {
    return &PostgresEvidenceRepository{DB: db}
}


func (repo *PostgresEvidenceRepository) GetEvidenceByCase(caseID string) ([]EvidenceDTO, error) {
    var evidences []EvidenceDTO
    result := repo.DB.Where("case_id = ?", caseID).Find(&evidences)
    return evidences, result.Error
}



func (repo *PostgresEvidenceRepository) GetEvidenceByID(evidenceID string) (*EvidenceDTO, error) {
    var evidence EvidenceDTO
    result := repo.DB.First(&evidence, "id = ?", evidenceID)
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, result.Error
    }
    return &evidence, nil
}


func (repo *PostgresEvidenceRepository) SearchEvidence(query string) ([]EvidenceResponse, error) {
    var evidences []EvidenceDTO
    pattern := "%" + query + "%"
    result := repo.DB.Where(
        "filename ILIKE ? OR file_type ILIKE ? OR metadata::text ILIKE ?",
        pattern, pattern, pattern,
    ).Find(&evidences)
    if result.Error != nil {
        return nil, result.Error
    }

    var responses []EvidenceResponse
    for _, ev := range evidences {
        responses = append(responses, EvidenceResponse{
            ID:         ev.ID,
            CaseID:     ev.CaseID,
            Filename:   ev.Filename,
            FileType:   ev.FileType,
            IPFSCID:    ev.IPFSCID,
            UploadedBy: ev.UploadedBy,
            Metadata:   ev.Metadata,
            UploadedAt: ev.UploadedAt.String(), // Or format with .Format("2006-01-02T15:04:05Z07:00")
        })
    }

    return responses, nil
}





func (repo *PostgresEvidenceRepository) GetFilteredEvidence(
    caseID string,
    filters map[string]interface{},
    sortField, sortOrder string,
) ([]EvidenceResponse, error) {
    var evidences []EvidenceDTO

    tx := repo.DB.Where("case_id = ?", caseID)

    for k, v := range filters {
        tx = tx.Where(fmt.Sprintf("%s = ?", k), v)
    }

    if sortField != "" && (sortOrder == "asc" || sortOrder == "desc") {
        tx = tx.Order(fmt.Sprintf("%s %s", sortField, sortOrder))
    }

    result := tx.Find(&evidences)
    if result.Error != nil {
        return nil, result.Error
    }

    //map evidenceDTO to EvidenceResponse
    var responses []EvidenceResponse
    for _, ev := range evidences {
        responses = append(responses, EvidenceResponse{
            ID:         ev.ID,
            CaseID:     ev.CaseID,
            Filename:   ev.Filename,
            FileType:   ev.FileType,
            IPFSCID:    ev.IPFSCID,
            UploadedBy: ev.UploadedBy,
            Metadata:   ev.Metadata,
            UploadedAt: ev.UploadedAt.String(),
        })
    }

    return responses, nil
}
