package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"martins-pocos/config"
	"martins-pocos/models"
	"martins-pocos/utils"

	"github.com/gorilla/mux"
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

	// Contar por status
	pendingCount := 0
	confirmedCount := 0
	completedCount := 0
	canceledCount := 0

	for _, req := range requests {
		switch req.Status {
		case "SOLICITADA":
			pendingCount++
		case "CONFIRMADA":
			confirmedCount++
		case "REALIZADA":
			completedCount++
		case "CANCELADA":
			canceledCount++
		}
	}

	data := struct {
		UserName       string
		Requests       []models.ServiceRequest
		PendingCount   int
		ConfirmedCount int
		CompletedCount int
		CanceledCount  int
		SuccessMsg     string
	}{
		UserName:       session.Values["user_name"].(string),
		Requests:       requests,
		PendingCount:   pendingCount,
		ConfirmedCount: confirmedCount,
		CompletedCount: completedCount,
		CanceledCount:  canceledCount,
	}

	if r.URL.Query().Get("success") == "created" {
		data.SuccessMsg = "Solicitação criada com sucesso!"
	} else if r.URL.Query().Get("success") == "updated" {
		data.SuccessMsg = "Solicitação atualizada com sucesso!"
	} else if r.URL.Query().Get("success") == "cancelled" {
		data.SuccessMsg = "Solicitação cancelada com sucesso!"
	}

	tmpl := template.Must(template.ParseFiles("templates/cliente_dashboard.html"))
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
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			utils.SendErrorResponse(w, "Erro ao processar formulário", http.StatusBadRequest)
			return
		}

		// Parse date
		preferredDate, err := time.Parse("2006-01-02", r.FormValue("preferred_date"))
		if err != nil {
			utils.SendErrorResponse(w, "Data inválida", http.StatusBadRequest)
			return
		}

		service := &models.ServiceRequest{
			UserID:        userID,
			FullName:      r.FormValue("full_name"),
			ServiceType:   r.FormValue("service_type"),
			Description:   r.FormValue("description"),
			CEP:           r.FormValue("cep"),
			Logradouro:    r.FormValue("logradouro"),
			Numero:        r.FormValue("numero"),
			Bairro:        r.FormValue("bairro"),
			Cidade:        r.FormValue("cidade"),
			Estado:        r.FormValue("estado"),
			PreferredDate: preferredDate,
			PreferredTime: r.FormValue("preferred_time"),
		}

		err = c.ServiceModel.Create(service)
		if err != nil {
			utils.SendErrorResponse(w, "Erro ao criar solicitação", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard/cliente?success=created", http.StatusFound)
	}
}

func (c *ServiceController) EditarSolicitacao(w http.ResponseWriter, r *http.Request) {
	session, _ := config.GetSessionStore().Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		// Buscar a solicitação
		service, err := c.ServiceModel.GetByIDAndUser(requestID, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Solicitação não encontrada", http.StatusNotFound)
				return
			}
			http.Error(w, "Erro ao buscar solicitação", http.StatusInternalServerError)
			return
		}

		// Verificar se pode editar (apenas SOLICITADA)
		if service.Status != "SOLICITADA" {
			http.Redirect(w, r, fmt.Sprintf("/solicitacao/%d", requestID), http.StatusFound)
			return
		}

		data := struct {
			Service *models.ServiceRequest
		}{
			Service: service,
		}

		tmpl := template.Must(template.ParseFiles("templates/editar_solicitacao.html"))
		tmpl.Execute(w, data)
		return
	}

	if r.Method == "POST" {
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			utils.SendErrorResponse(w, "Erro ao processar formulário", http.StatusBadRequest)
			return
		}

		// Parse date
		preferredDate, err := time.Parse("2006-01-02", r.FormValue("preferred_date"))
		if err != nil {
			utils.SendErrorResponse(w, "Data inválida", http.StatusBadRequest)
			return
		}

		service := &models.ServiceRequest{
			ID:            requestID,
			UserID:        userID,
			FullName:      r.FormValue("full_name"),
			ServiceType:   r.FormValue("service_type"),
			Description:   r.FormValue("description"),
			CEP:           r.FormValue("cep"),
			Logradouro:    r.FormValue("logradouro"),
			Numero:        r.FormValue("numero"),
			Bairro:        r.FormValue("bairro"),
			Cidade:        r.FormValue("cidade"),
			Estado:        r.FormValue("estado"),
			PreferredDate: preferredDate,
			PreferredTime: r.FormValue("preferred_time"),
		}

		err = c.ServiceModel.Update(service)
		if err != nil {
			if err == sql.ErrNoRows {
				utils.SendErrorResponse(w, "Solicitação não pode ser editada", http.StatusBadRequest)
				return
			}
			utils.SendErrorResponse(w, "Erro ao atualizar solicitação", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard/cliente?success=updated", http.StatusFound)
	}
}

func (c *ServiceController) VerSolicitacao(w http.ResponseWriter, r *http.Request) {
	session, _ := config.GetSessionStore().Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// Buscar a solicitação
	service, err := c.ServiceModel.GetByIDAndUser(requestID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Solicitação não encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "Erro ao buscar solicitação", http.StatusInternalServerError)
		return
	}

	data := struct {
		Service *models.ServiceRequest
	}{
		Service: service,
	}

	tmpl := template.Must(template.ParseFiles("templates/ver_solicitacao.html"))
	tmpl.Execute(w, data)
}

func (c *ServiceController) CancelarSolicitacao(w http.ResponseWriter, r *http.Request) {
	session, _ := config.GetSessionStore().Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.SendErrorResponse(w, "ID inválido", http.StatusBadRequest)
		return
	}

	err = c.ServiceModel.Cancel(requestID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.SendErrorResponse(w, "Solicitação não pode ser cancelada", http.StatusBadRequest)
			return
		}
		utils.SendErrorResponse(w, "Erro ao cancelar solicitação", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard/cliente?success=cancelled", http.StatusFound)
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
	validStatuses := map[string]bool{
		"SOLICITADA": true,
		"CONFIRMADA": true,
		"REALIZADA":  true,
		"CANCELADA":  true,
	}

	if !validStatuses[status] {
		http.Error(w, "Status inválido", http.StatusBadRequest)
		return
	}

	err = c.ServiceModel.UpdateStatus(requestID, status)
	if err != nil {
		http.Error(w, "Erro ao atualizar status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}