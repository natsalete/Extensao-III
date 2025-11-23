package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"martins-pocos/config"
	"martins-pocos/routes"

	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis de ambiente do arquivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Aviso: Arquivo .env não encontrado, usando variáveis do sistema")
	}

	// Verificar credenciais do WhatsApp
	apiKey := os.Getenv("WHATSAPP_API_KEY")
	instanceId := os.Getenv("WHATSAPP_INSTANCE_ID")
	
	if apiKey != "" && instanceId != "" {
		log.Printf("✅ WhatsApp configurado")
		log.Printf("   🔑 API Key: %s", maskString(apiKey))
		log.Printf("   🆔 Instance: %s", maskString(instanceId))
	} else {
		log.Println("⚠️ WhatsApp não configurado (as notificações não funcionarão)")
	}

	// Initialize database
	config.InitDB()
	defer config.GetDB().Close()

	// Initialize session store
	config.InitSession()

	// Setup routes
	r := routes.SetupRoutes()

	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("🚀 Servidor Martins Poços")
	fmt.Println("========================================")
	fmt.Println("📍 Porta: 8090")
	fmt.Println("🌐 URL: http://localhost:8090")
	fmt.Println("========================================")
	fmt.Println("")
	
	log.Fatal(http.ListenAndServe(":8090", r))
}

// maskString mascara strings sensíveis para logs
func maskString(s string) string {
	if len(s) < 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}