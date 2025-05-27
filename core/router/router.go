package router

// import (
// 	"log"
// 	"net/http"
//  	"aegis-api/services/registration"
// 	"aegis-api/services/login/auth"
// 	"aegis-api/services/case_creation"

// 	"aegis-api/db"

// )

// func StartServer() {
	
		
// repo := registration.NewGormUserRepository(db.DB)
// service := registration.NewRegistrationService(repo)
// 	http.HandleFunc("/register", registration.RegisterHandler(service))
// 	http.HandleFunc("/verify", registration.VerifyHandler(repo))

// 	http.HandleFunc("/login", auth.LoginHandler)
// 	http.HandleFunc("/cases", case_creation.CreateCase)

// 	log.Println("Server running on http://localhost:8080")
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }
