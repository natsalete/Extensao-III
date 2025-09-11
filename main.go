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

	fmt.Println("Servidor rodando na porta 8080")
	fmt.Println("Acesse: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}