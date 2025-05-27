package user_management

import (
    "github.com/google/uuid"
)

type User struct {
    ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
    FullName string
    Email    string
    Role     string `gorm:"type:user_role"`
    // ...
}
