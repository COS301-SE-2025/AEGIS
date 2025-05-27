// case_creation/types.go

package case_assign

import (
  "time"
  "github.com/google/uuid"
)

type CaseUserRole struct {
  UserID     uuid.UUID
  CaseID     uuid.UUID
  Role       string    // One of your user_role ENUM values
  AssignedAt time.Time
}
