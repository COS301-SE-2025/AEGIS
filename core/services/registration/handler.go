package registration

import (
	"encoding/json"
	"net/http"

)

// RegisterHandler handles HTTP POST requests for user registration.
func RegisterHandler(service *RegistrationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow only POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON request
		// Decode request
		var req RegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Validate input
		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}


		// Call service to register the user
		userEntity, err := service.Register(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		// Return UserResponse as JSON
		response := EntityToResponse(userEntity)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
func VerifyHandler(repo UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Missing verification token", http.StatusBadRequest)
			return
		}

		user, err := repo.GetUserByToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusNotFound)
			return
		}

		user.IsVerified = true
		user.VerificationToken = ""
		repo.UpdateUser(user)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(" Email verified successfully"))
	}
}
