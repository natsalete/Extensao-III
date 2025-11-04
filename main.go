package main

import (
	"fmt"
	"log"
	"net/http"

	"martins-pocos/config"
	"martins-pocos/routes"
)

func main() {
	// Initialize database
	config.InitDB()
	defer config.GetDB().Close()

	// Initialize session store
	config.InitSession()

	// Setup routes
	r := routes.SetupRoutes()

	fmt.Println("Servidor rodando na porta 8090")
	fmt.Println("Acesse: http://localhost:8090")
	log.Fatal(http.ListenAndServe(":8090", r))
}