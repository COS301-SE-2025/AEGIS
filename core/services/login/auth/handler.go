package auth

// import (
//     "encoding/json"
//     "net/http"
// )

// func LoginHandler(w http.ResponseWriter, r *http.Request) {
//     var req LoginRequest
//     if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//         http.Error(w, "invalid request", http.StatusBadRequest)
//         return
//     }

//     resp, err := Login(req.Email, req.Password)
//     if err != nil {
//         http.Error(w, err.Error(), http.StatusUnauthorized)
//         return
//     }

//     w.Header().Set("Content-Type", "application/json")
//     json.NewEncoder(w).Encode(resp)
// }
