package case_creation   

type CreateCaseRequest struct {
    Title              string `json:"title" validate:"required"`
    Description        string `json:"description"`
    Status             string `json:"status"` // optional: default is handled by DB
    Priority           string `json:"priority"`
    InvestigationStage string `json:"investigation_stage"`
    CreatedBy          string `json:"created_by" validate:"required,uuid"`
    TeamName           string `json:"team_name" validate:"required"`
}


// // CreateCaseRequest represents the payload to create a new case.
// type CreateCaseRequest struct {
//     Title              string ` + "`json:"title" validate:"required"`" + `
//     Description        string ` + "`json:"description"`" + `
//     Status             string ` + "`json:"status"`" + ` // optional: default is handled by DB
//     Priority           string ` + "`json:"priority"`" + `
//     InvestigationStage string ` + "`json:"investigation_stage"`" + `
//     CreatedBy          string ` + "`json:"created_by" validate:"required,uuid"`" + `
//     TeamName           string ` + "`json:"team_name" validate:"required"`" + `
// }
