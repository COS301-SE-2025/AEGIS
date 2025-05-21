package main    

import (
	"log"
	"net/http"
	"aegis-registration/services/registration"
    "aegis-backend/router"
)
func main() {

   db := registration.InitDB()
	repo := registration.NewGormUserRepository(db)
	service := registration.NewRegistrationService(repo)

router.StartServer()

}
