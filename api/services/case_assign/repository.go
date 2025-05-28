package case_assign

import (
  //"aegis-api/db"
 
  "github.com/google/uuid"
  "gorm.io/gorm"
   
)

type GormCaseAssignmentRepo struct {
  db *gorm.DB
}

func NewGormCaseAssignmentRepo(database *gorm.DB) *GormCaseAssignmentRepo {
  return &GormCaseAssignmentRepo{db: database}
}

func (r *GormCaseAssignmentRepo) AssignRole(userID, caseID uuid.UUID, role string) error {
  entry := CaseUserRole{
    UserID: userID,
    CaseID: caseID,
    Role:   role,
  }
  return r.db.Create(&entry).Error
}

func (r *GormCaseAssignmentRepo) ListAssignments(caseID uuid.UUID) ([]CaseUserRole, error) {
  var assignments []CaseUserRole
  err := r.db.Where("case_id = ?", caseID).Find(&assignments).Error
  return assignments, err
}

func (r *GormCaseAssignmentRepo) UnassignRole(userID, caseID uuid.UUID) error {
  return r.db.Where("user_id = ? AND case_id = ?", userID, caseID).Delete(&CaseUserRole{}).Error
}

// func (r *GormCaseAssignmentRepo) IsAdmin(userID uuid.UUID) (bool, error) {
//   type result struct {
//     IsAdmin bool
//   }
//   var out result
//   err := r.db.Table("users").Select("is_admin").Where("id = ?", userID).Scan(&out).Error
//   return out.IsAdmin, err
// }
func (r *GormCaseAssignmentRepo) IsAdmin(userID uuid.UUID) (bool, error) {
    type result struct {
        Role string
    }
    var out result
    err := r.db.Table("users").
        Select("role").
        Where("id = ?", userID).
        Scan(&out).Error
    if err != nil {
        return false, err
    }
    return out.Role == "Admin", nil
}

