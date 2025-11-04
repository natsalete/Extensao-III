package controllers

import (
	"html/template"
	"net/http"
	"strconv"

	"martins-pocos/config"
	"martins-pocos/models"

	"github.com/gorilla/mux"
)

type ServiceController struct {
	ServiceModel *models.ServiceModel
}

func NewServiceController(serviceModel *models.ServiceModel) *ServiceController {
	return &ServiceController{ServiceModel: serviceModel}
}

// ClienteDashboard - Dashboard do cliente
func (c *ServiceController) ClienteDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := c.getUserID(w, r)
	if !ok {
		return
	}

	requests, err := c.ServiceModel.GetByUserID(userID)
	if err != nil {
		http.Error(w, "Erro ao buscar solicitações", http.StatusInternalServerError)
		return
	}

	session, _ := config.GetSessionStore().Get(r, "session")
	
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
		PendingCount:   c.countByStatus(requests, 1), // SOLICITADA
		ConfirmedCount: c.countByStatus(requests, 2), // CONFIRMADA
		CompletedCount: c.countByStatus(requests, 3), // REALIZADA
		CanceledCount:  c.countByStatus(requests, 4), // CANCELADA
		SuccessMsg:     c.getSuccessMessage(r),
	}

	c.renderTemplate(w, "templates/cliente_dashboard.html", data)
}

// SolicitarServico - Criar nova solicitação
func (c *ServiceController) SolicitarServico(w http.ResponseWriter, r *http.Request) {
	userID, ok := c.getUserID(w, r)
	if !ok {
		return
	}

	if r.Method == "GET" {
		c.showServiceRequestForm(w)
		return
	}

	c.createServiceRequest(w, r, userID)
}

// EditarSolicitacao - Editar solicitação existente
func (c *ServiceController) EditarSolicitacao(w http.ResponseWriter, r *http.Request) {
	userID, ok := c.getUserID(w, r)
	if !ok {
		return
	}

	requestID, err := c.getRequestID(r)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		c.showEditForm(w, r, requestID, userID)
		return
	}

	c.updateServiceRequest(w, r, requestID, userID)
}

// VerSolicitacao - Ver detalhes da solicitação
func (c *ServiceController) VerSolicitacao(w http.ResponseWriter, r *http.Request) {
	userID, ok := c.getUserID(w, r)
	if !ok {
		return
	}

	requestID, err := c.getRequestID(r)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	service, err := c.ServiceModel.GetByIDAndUser(requestID, userID)
	if err != nil {
		c.handleNotFound(w, err, "Solicitação não encontrada")
		return
	}

	data := struct {
		Service *models.ServiceRequest
	}{
		Service: service,
	}

	c.renderTemplate(w, "templates/ver_solicitacao.html", data)
}

// CancelarSolicitacao - Cancelar solicitação
func (c *ServiceController) CancelarSolicitacao(w http.ResponseWriter, r *http.Request) {
	userID, ok := c.getUserID(w, r)
	if !ok {
		return
	}

	requestID, err := c.getRequestID(r)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := c.ServiceModel.Cancel(requestID, userID); err != nil {
		c.handleUpdateError(w, err, "Solicitação não pode ser cancelada")
		return
	}

	http.Redirect(w, r, "/dashboard/cliente?success=cancelled", http.StatusFound)
}

// Helper methods
func (c *ServiceController) getUserID(w http.ResponseWriter, r *http.Request) (int, bool) {
	session, _ := config.GetSessionStore().Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return 0, false
	}
	return userID, true
}

func (c *ServiceController) getRequestID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["id"])
}

func (c *ServiceController) countByStatus(requests []models.ServiceRequest, statusID int) int {
	count := 0
	for _, req := range requests {
		if req.StatusID == statusID {
			count++
		}
	}
	return count
}

func (c *ServiceController) getSuccessMessage(r *http.Request) string {
	switch r.URL.Query().Get("success") {
	case "created":
		return "Solicitação criada com sucesso!"
	case "updated":
		return "Solicitação atualizada com sucesso!"
	case "cancelled":
		return "Solicitação cancelada com sucesso!"
	default:
		return ""
	}
}

func (c *ServiceController) renderTemplate(w http.ResponseWriter, templatePath string, data interface{}) {
	tmpl := template.Must(template.ParseFiles(templatePath))
	tmpl.Execute(w, data)
}

func (c *ServiceController) handleNotFound(w http.ResponseWriter, err error, message string) {
	if err.Error() == "sql: no rows in result set" {
		http.Error(w, message, http.StatusNotFound)
		return
	}
	http.Error(w, "Erro ao buscar dados", http.StatusInternalServerError)
}

func (c *ServiceController) handleUpdateError(w http.ResponseWriter, err error, message string) {
	if err.Error() == "sql: no rows in result set" {
		http.Error(w, message, http.StatusBadRequest)
		return
	}
	http.Error(w, "Erro ao processar solicitação", http.StatusInternalServerError)
}

func (c *ServiceController) showServiceRequestForm(w http.ResponseWriter) {
	serviceTypes, err := c.ServiceModel.GetAllServiceTypes()
	if err != nil {
		http.Error(w, "Erro ao carregar tipos de serviço", http.StatusInternalServerError)
		return
	}

	data := struct {
		ServiceTypes []models.ServiceType
	}{
		ServiceTypes: serviceTypes,
	}

	c.renderTemplate(w, "templates/solicitar_servico.html", data)
}

func (c *ServiceController) showEditForm(w http.ResponseWriter, r *http.Request, requestID, userID int) {
	service, err := c.ServiceModel.GetByIDAndUser(requestID, userID)
	if err != nil {
		c.handleNotFound(w, err, "Solicitação não encontrada")
		return
	}

	// Apenas solicitações com status SOLICITADA (id=1) podem ser editadas
	if service.StatusID != 1 {
		http.Redirect(w, r, "/solicitacao/"+strconv.Itoa(requestID), http.StatusFound)
		return
	}

	serviceTypes, err := c.ServiceModel.GetAllServiceTypes()
	if err != nil {
		http.Error(w, "Erro ao carregar tipos de serviço", http.StatusInternalServerError)
		return
	}

	data := struct {
		Service      *models.ServiceRequest
		ServiceTypes []models.ServiceType
	}{
		Service:      service,
		ServiceTypes: serviceTypes,
	}

	c.renderTemplate(w, "templates/editar_solicitacao.html", data)
}