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

	// Initialize services
	whatsappService := services.NewWhatsAppService()

	// Initialize controllers
	homeController := controllers.NewHomeController()
	authController := controllers.NewAuthController(userModel)
	serviceController := controllers.NewServiceController(serviceModel)
	adminController := controllers.NewAdminController(serviceModel, whatsappService)

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Public routes
	r.HandleFunc("/", homeController.Index)
	r.HandleFunc("/login", authController.LoginPage).Methods("GET")
	r.HandleFunc("/login", authController.Login).Methods("POST")
	r.HandleFunc("/register", authController.RegisterPage).Methods("GET")
	r.HandleFunc("/register", authController.Register).Methods("POST")

	// ========== CLIENT ROUTES (Protected) ==========
	r.HandleFunc("/dashboard/cliente", 
		middleware.RequireAuth(middleware.RequireClient(serviceController.ClienteDashboard))).Methods("GET")
	
	// Service request management
	r.HandleFunc("/solicitar-servico", 
		middleware.RequireAuth(middleware.RequireClient(serviceController.SolicitarServico))).Methods("GET", "POST")
	r.HandleFunc("/solicitacao/{id:[0-9]+}", 
		middleware.RequireAuth(middleware.RequireClient(serviceController.VerSolicitacao))).Methods("GET")
	r.HandleFunc("/solicitacao/{id:[0-9]+}/editar", 
		middleware.RequireAuth(middleware.RequireClient(serviceController.EditarSolicitacao))).Methods("GET", "POST")
	r.HandleFunc("/solicitacao/{id:[0-9]+}/cancelar", 
		middleware.RequireAuth(middleware.RequireClient(serviceController.CancelarSolicitacao))).Methods("POST")
	
	// ========== ADMIN ROUTES (Protected + Admin Only) ==========
	r.HandleFunc("/dashboard/admin", 
		middleware.RequireAuth(middleware.RequireAdmin(adminController.AdminDashboard))).Methods("GET")
	
	// Admin service request management
	r.HandleFunc("/admin/solicitacao/{id:[0-9]+}", 
		middleware.RequireAuth(middleware.RequireAdmin(adminController.VerSolicitacaoAdmin))).Methods("GET")
	r.HandleFunc("/admin/solicitacao/{id:[0-9]+}/editar", 
		middleware.RequireAuth(middleware.RequireAdmin(adminController.EditarSolicitacaoAdmin))).Methods("GET", "POST")
	r.HandleFunc("/admin/solicitacao/{id:[0-9]+}/deletar", 
		middleware.RequireAuth(middleware.RequireAdmin(adminController.DeletarSolicitacao))).Methods("POST")
	
	// Status update
	r.HandleFunc("/admin/update-status", 
		middleware.RequireAuth(middleware.RequireAdmin(adminController.UpdateStatus))).Methods("POST")
	
	// ========== AUTH ROUTES ==========
	r.HandleFunc("/logout", middleware.RequireAuth(authController.Logout))

	return r
}