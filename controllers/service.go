package controllers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"martins-pocos/config"
	"martins-pocos/models"
)

type ServiceController struct {
	ServiceModel *models.ServiceModel
}

func NewServiceController(serviceModel *models.ServiceModel) *ServiceController {
	return &ServiceController{ServiceModel: serviceModel}
}

func (c *ServiceController) ClienteDashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := config.GetSessionStore().Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	requests, err := c.ServiceModel.GetByUserID(userID)
	if err != nil {
		http.Error(w, "Erro ao buscar solicitações", http.StatusInternalServerError)
		return
	}

	data := struct {
		UserName string
		Requests []models.ServiceRequest
	}{
		UserName: session.Values["user_name"].(string),
		Requests: requests,
	}

	tmpl := template.Must(template.ParseFiles("templates/cliente_dashboard.html"))
	tmpl.Execute(w, data)
}

func (c *ServiceController) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := config.GetSessionStore().Get(r, "session")
	userType, ok := session.Values["user_type"].(string)
	if !ok || userType != "gestor" {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	requests, err := c.ServiceModel.GetAll()
	if err != nil {
		http.Error(w, "Erro ao buscar solicitações", http.StatusInternalServerError)
		return
	}

	data := struct {
		UserName string
		Requests []models.ServiceRequest
	}{
		UserName: session.Values["user_name"].(string),
		Requests: requests,
	}

	tmpl := template.Must(template.ParseFiles("templates/admin_dashboard.html"))
	tmpl.Execute(w, data)
}

func (c *ServiceController) SolicitarServico(w http.ResponseWriter, r *http.Request) {
	session, _ := config.GetSessionStore().Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/solicitar_servico.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		service := &models.ServiceRequest{
			UserID:      userID,
			ServiceType: r.FormValue("service_type"),
			Description: r.FormValue("description"),
		}

		err := c.ServiceModel.Create(service)
		if err != nil {
			http.Error(w, "Erro ao criar solicitação", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard/cliente?success=1", http.StatusFound)
	}
}

func (c *ServiceController) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	session, _ := config.GetSessionStore().Get(r, "session")
	userType, ok := session.Values["user_type"].(string)
	if !ok || userType != "gestor" {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	requestID, err := strconv.Atoi(r.FormValue("request_id"))
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")

	err = c.ServiceModel.UpdateStatus(requestID, status)
	if err != nil {
		http.Error(w, "Erro ao atualizar status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}