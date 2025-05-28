package Evidence_Viewer

import (
    "database/sql"
    "aegis-api/models"
	"fmt"
	
)

type SQLDB interface {
    Query(query string, args ...any) (*sql.Rows, error)
    QueryRow(query string, args ...any) *sql.Row
}

type PostgresEvidenceRepository struct {
    DB SQLDB
}

func (repo *PostgresEvidenceRepository) GetEvidenceByCase(caseID string) ([]models.EvidenceDTO, error) {
	//sql query with sql injection prevention
    rows, err := repo.DB.Query(`
        SELECT id, case_id, uploaded_by, filename, file_type, ipfs_cid, file_size, checksum, metadata, uploaded_at 
        FROM evidence WHERE case_id = $1`, caseID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var evidences []models.EvidenceDTO
    for rows.Next() {
        var ev models.EvidenceDTO
		//extarct column values 
        err := rows.Scan(&ev.ID, &ev.CaseID, &ev.UploadedBy, &ev.Filename, &ev.FileType, &ev.IPFSCID, &ev.FileSize, &ev.Checksum, &ev.Metadata, &ev.UploadedAt)
        if err != nil {
            return nil, err
        }
		//store values in dto
        evidences = append(evidences, ev)
    }

    return evidences, nil
}


func (repo *PostgresEvidenceRepository) GetEvidenceByID(evidenceID string) (*models.EvidenceDTO, error) {
    var ev models.EvidenceDTO
    err := repo.DB.QueryRow(`
        SELECT id, case_id, uploaded_by, filename, file_type, ipfs_cid, file_size, checksum, metadata, uploaded_at 
        FROM evidence WHERE id = $1`, evidenceID).
        Scan(&ev.ID, &ev.CaseID, &ev.UploadedBy, &ev.Filename, &ev.FileType, &ev.IPFSCID,&ev.FileSize, &ev.Checksum, &ev.Metadata, &ev.UploadedAt)
    
    if err == sql.ErrNoRows {
        return nil, nil
    } else if err != nil {
        return nil, err
    }

    return &ev, nil
}




func (repo *PostgresEvidenceRepository) SearchEvidence(query string) ([]models.EvidenceResponse, error) {
    q := "%" + query + "%"
    rows, err := repo.DB.Query(`
        SELECT id, filename, file_type, ipfs_cid 
        FROM evidence 
        WHERE filename ILIKE $1 OR file_type ILIKE $1 OR metadata::text ILIKE $1`, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []models.EvidenceResponse
    for rows.Next() {
        var res models.EvidenceResponse
        err := rows.Scan(&res.ID, &res.Filename, &res.FileType, &res.IPFSCID)
        if err != nil {
            return nil, err
        }
        results = append(results, res)
    }

    return results, nil
}




func (repo *PostgresEvidenceRepository) GetFilteredEvidence(caseID string, filters map[string]interface{}, sortField string, sortOrder string) ([]models.EvidenceResponse, error) {
    baseQuery := `
        SELECT id, filename, file_type, ipfs_cid
        FROM evidence
        WHERE case_id = $1`
    
    args := []interface{}{caseID}
    argIndex := 2

    for k, v := range filters {
        baseQuery += fmt.Sprintf(" AND %s = $%d", k, argIndex)
        args = append(args, v)
        argIndex++
    }

    if sortField != "" && (sortOrder == "asc" || sortOrder == "desc") {
        baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder)
    }

    rows, err := repo.DB.Query(baseQuery, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []models.EvidenceResponse
    for rows.Next() {
        var res models.EvidenceResponse
        err := rows.Scan(&res.ID, &res.Filename, &res.FileType, &res.IPFSCID)
        if err != nil {
            return nil, err
        }
        results = append(results, res)
    }

    return results, nil
}
