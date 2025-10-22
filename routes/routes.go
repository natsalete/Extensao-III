package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"martins-pocos/config"
	"martins-pocos/controllers"
	"martins-pocos/middleware"
	"martins-pocos/models"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Initialize models
	userModel := models.NewUserModel(config.GetDB())
	serviceModel := models.NewServiceModel(config.GetDB())

	// Initialize controllers
	homeController := controllers.NewHomeController()
	authController := controllers.NewAuthController(userModel)
	serviceController := controllers.NewServiceController(serviceModel)
	// adminController := controllers.NewAdminController(serviceModel)

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Public routes
	r.HandleFunc("/", homeController.Index)
	r.HandleFunc("/login", authController.LoginPage).Methods("GET")
	r.HandleFunc("/login", authController.Login).Methods("POST")
	r.HandleFunc("/register", authController.RegisterPage).Methods("GET")
	r.HandleFunc("/register", authController.Register).Methods("POST")

	// ========== CLIENT ROUTES (Protected) ==========
	r.HandleFunc("/dashboard/cliente", middleware.RequireAuth(serviceController.ClienteDashboard)).Methods("GET")
	
	// Service request management
	r.HandleFunc("/solicitar-servico", middleware.RequireAuth(serviceController.SolicitarServico)).Methods("GET", "POST")
	r.HandleFunc("/solicitacao/{id:[0-9]+}", middleware.RequireAuth(serviceController.VerSolicitacao)).Methods("GET")
	r.HandleFunc("/solicitacao/{id:[0-9]+}/editar", middleware.RequireAuth(serviceController.EditarSolicitacao)).Methods("GET", "POST")
	r.HandleFunc("/solicitacao/{id:[0-9]+}/cancelar", middleware.RequireAuth(serviceController.CancelarSolicitacao)).Methods("POST")
	
	// ========== ADMIN ROUTES (Protected + Admin Only) ==========
	// r.HandleFunc("/dashboard/admin", middleware.RequireAuth(middleware.RequireAdmin(adminController.AdminDashboard))).Methods("GET")
	// r.HandleFunc("/admin/update-status", middleware.RequireAuth(middleware.RequireAdmin(adminController.UpdateStatus))).Methods("POST")
	
	// ========== AUTH ROUTES ==========
	r.HandleFunc("/logout", middleware.RequireAuth(authController.Logout))

	return r
}