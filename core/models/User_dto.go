package models


type UserDTO struct {
    ID           string `json:"id"`
    FullName     string `json:"full_name"`
    Email        string `json:"email"`
    Role         string `json:"role"`  
    IsVerified   bool   `json:"is_verified"`
    CreatedAt    string `json:"created_at"`
}