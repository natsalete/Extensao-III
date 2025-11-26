package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"martins-pocos/config"
	"martins-pocos/controllers"
	"martins-pocos/middleware"
	"martins-pocos/models"
	"martins-pocos/services"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Initialize models
	userModel := models.NewUserModel(config.GetDB())
	serviceModel := models.NewServiceModel(config.GetDB())
	contractModel := models.NewContractModel(config.GetDB())

	// Initialize services
	whatsappService := services.NewWhatsAppService()

	// Initialize controllers
	homeController := controllers.NewHomeController()
	authController := controllers.NewAuthController(userModel)
	serviceController := controllers.NewServiceController(serviceModel)
	adminController := controllers.NewAdminController(serviceModel, whatsappService)
	contractController := controllers.NewContractController(contractModel, serviceModel)

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// ========== PUBLIC ROUTES ==========
	r.HandleFunc("/", homeController.Index)
	r.HandleFunc("/login", authController.LoginPage).Methods("GET")
	r.HandleFunc("/login", authController.Login).Methods("POST")
	r.HandleFunc("/register", authController.RegisterPage).Methods("GET")
	r.HandleFunc("/register", authController.Register).Methods("POST")
	r.HandleFunc("/logout", middleware.RequireAuth(authController.Logout))

	// ========== ADMIN ROUTES (Protected + Admin Only) ==========
	// IMPORTANTE: Rotas ADMIN devem vir ANTES das rotas CLIENT para evitar conflitos!
	
	// Dashboard Admin
	r.HandleFunc("/dashboard/admin", 
		middleware.RequireAuth(middleware.RequireAdmin(adminController.AdminDashboard))).Methods("GET")
	
	// Admin service request management
	r.HandleFunc("/admin/solicitacao/{id:[0-9]+}", 
		middleware.RequireAuth(middleware.RequireAdmin(adminController.VerSolicitacaoAdmin))).Methods("GET")
	r.HandleFunc("/admin/solicitacao/{id:[0-9]+}/editar", 
		middleware.RequireAuth(middleware.RequireAdmin(adminController.EditarSolicitacaoAdmin))).Methods("GET", "POST")
	r.HandleFunc("/admin/solicitacao/{id:[0-9]+}/deletar", 
		middleware.RequireAuth(middleware.RequireAdmin(adminController.DeletarSolicitacao))).Methods("POST")
	r.HandleFunc("/admin/solicitacao/{id:[0-9]+}/criar-contrato", 
		middleware.RequireAuth(middleware.RequireAdmin(contractController.CreateContract))).Methods("GET", "POST")
	
	// Status update
	r.HandleFunc("/admin/update-status", 
		middleware.RequireAuth(middleware.RequireAdmin(adminController.UpdateStatus))).Methods("POST")
	
	// Admin contract routes - ORDEM IMPORTANTE!
	// Lista deve vir ANTES dos detalhes com {id}
	r.HandleFunc("/admin/contratos", 
		middleware.RequireAuth(middleware.RequireAdmin(contractController.ListContracts))).Methods("GET")
	r.HandleFunc("/admin/contratos/{id:[0-9]+}/editar", 
		middleware.RequireAuth(middleware.RequireAdmin(contractController.EditContract))).Methods("GET", "POST")
	r.HandleFunc("/admin/contratos/{id:[0-9]+}/enviar-assinatura", 
		middleware.RequireAuth(middleware.RequireAdmin(contractController.SendForSignature))).Methods("POST")
	r.HandleFunc("/admin/contratos/{id:[0-9]+}/assinar-empresa", 
		middleware.RequireAuth(middleware.RequireAdmin(contractController.SignContractCompany))).Methods("POST")
	r.HandleFunc("/admin/contratos/{id:[0-9]+}", 
		middleware.RequireAuth(middleware.RequireAdmin(contractController.ViewContract))).Methods("GET")
	
	// ========== CLIENT ROUTES (Protected) ==========
	// Rotas CLIENT vêm DEPOIS das rotas ADMIN
	
	// Dashboard Cliente
	r.HandleFunc("/dashboard/cliente", 
		middleware.RequireAuth(middleware.RequireClient(serviceController.ClienteDashboard))).Methods("GET")
	
	// Service request management
	r.HandleFunc("/solicitar-servico", 
		middleware.RequireAuth(middleware.RequireClient(serviceController.SolicitarServico))).Methods("GET", "POST")
	r.HandleFunc("/solicitacao/{id:[0-9]+}/editar", 
		middleware.RequireAuth(middleware.RequireClient(serviceController.EditarSolicitacao))).Methods("GET", "POST")
	r.HandleFunc("/solicitacao/{id:[0-9]+}/cancelar", 
		middleware.RequireAuth(middleware.RequireClient(serviceController.CancelarSolicitacao))).Methods("POST")
	r.HandleFunc("/solicitacao/{id:[0-9]+}", 
		middleware.RequireAuth(middleware.RequireClient(serviceController.VerSolicitacao))).Methods("GET")
	
	// Client contract routes - ORDEM IMPORTANTE!
	r.HandleFunc("/contratos", 
		middleware.RequireAuth(middleware.RequireClient(contractController.ClientContracts))).Methods("GET")
	r.HandleFunc("/contratos/{id:[0-9]+}/assinar", 
		middleware.RequireAuth(middleware.RequireClient(contractController.ClientSignContract))).Methods("POST")
	r.HandleFunc("/contratos/{id:[0-9]+}", 
		middleware.RequireAuth(middleware.RequireClient(contractController.ClientViewContract))).Methods("GET")

	return r
}