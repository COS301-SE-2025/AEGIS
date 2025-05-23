// package case_creation

// import (
// 	"encoding/json"
// 	"io/ioutil"
// 	"net/http"
// 	"github.com/google/uuid"
// 	"aegis-api/db"
// )


// func CreateCase(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Unable to read request", http.StatusBadRequest)
// 		return
// 	}
// 	defer r.Body.Close()

// 	var req CreateCaseRequest
// 	if err := json.Unmarshal(body, &req); err != nil {
// 		http.Error(w, "Invalid JSON", http.StatusBadRequest)
// 		return
// 	}

// 	createdByUUID, err := uuid.Parse(req.CreatedBy)
// 	if err != nil {
// 		http.Error(w, "Invalid created_by UUID", http.StatusBadRequest)
// 		return
// 	}

// 	newCase := Case{
// 		ID:                 uuid.New(),
// 		Title:              req.Title,
// 		Description:        req.Description,
// 		Status:             req.Status,
// 		Priority:           req.Priority,
// 		InvestigationStage: req.InvestigationStage,
// 		CreatedBy:          createdByUUID,
// 	}

// 	if err := db.DB.Create(&newCase).Error; err != nil {
// 		http.Error(w, "Database error", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(newCase)
// }
