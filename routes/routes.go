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

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Public routes
	r.HandleFunc("/", homeController.Index)
	r.HandleFunc("/login", authController.LoginPage).Methods("GET")
	r.HandleFunc("/login", authController.Login).Methods("POST")
	r.HandleFunc("/register", authController.RegisterPage).Methods("GET")
	r.HandleFunc("/register", authController.Register).Methods("POST")

	// Protected routes with authentication middleware
	r.HandleFunc("/dashboard/cliente", middleware.RequireAuth(serviceController.ClienteDashboard))
	r.HandleFunc("/dashboard/admin", middleware.RequireAuth(middleware.RequireAdmin(serviceController.AdminDashboard)))
	r.HandleFunc("/solicitar-servico", middleware.RequireAuth(serviceController.SolicitarServico))
	r.HandleFunc("/update-status", middleware.RequireAuth(middleware.RequireAdmin(serviceController.UpdateStatus))).Methods("POST")
	r.HandleFunc("/logout", middleware.RequireAuth(authController.Logout))

	return r
}