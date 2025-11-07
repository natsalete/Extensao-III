package controllers

import (
	// "fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

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

// PageData estrutura comum para todas as páginas
type PageData struct {
	PageTitle         string
	UserName          string
	CustomCSS         string
	CustomJS          string
	AdditionalScripts []string
	CurrentYear       int
	Data              interface{}
}

// ClienteDashboard - Dashboard do cliente com filtros e paginação
func (c *ServiceController) ClienteDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := c.getUserID(w, r)
	if !ok {
		return
	}

	// Capturar parâmetros de filtro e paginação
	statusFilter := r.URL.Query().Get("status")
	serviceTypeFilter := r.URL.Query().Get("service_type")
	pageStr := r.URL.Query().Get("page")
	
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	pageSize := 5
	offset := (page - 1) * pageSize

	// Buscar solicitações com filtros
	requests, totalCount, err := c.ServiceModel.GetByUserIDWithFilters(
		userID, 
		statusFilter, 
		serviceTypeFilter, 
		pageSize, 
		offset,
	)
	if err != nil {
		http.Error(w, "Erro ao buscar solicitações", http.StatusInternalServerError)
		return
	}

	// Calcular total de páginas
	totalPages := (totalCount + pageSize - 1) / pageSize

	// Buscar tipos de serviço para o filtro
	serviceTypes, err := c.ServiceModel.GetAllServiceTypes()
	if err != nil {
		http.Error(w, "Erro ao carregar tipos de serviço", http.StatusInternalServerError)
		return
	}

	// Buscar status para o filtro
	statuses, err := c.ServiceModel.GetAllStatuses()
	if err != nil {
		http.Error(w, "Erro ao carregar status", http.StatusInternalServerError)
		return
	}

	// Contar por status (total sem filtros para os cards)
	allRequests, err := c.ServiceModel.GetByUserID(userID)
	if err != nil {
		allRequests = []models.ServiceRequest{}
	}

	session, _ := config.GetSessionStore().Get(r, "session")
	userName := session.Values["user_name"].(string)

	data := struct {
		UserName          string
		Requests          []models.ServiceRequest
		ServiceTypes      []models.ServiceType
		Statuses          []models.RequestStatus
		PendingCount      int
		ConfirmedCount    int
		CompletedCount    int
		CanceledCount     int
		CurrentPage       int
		TotalPages        int
		TotalCount        int
		HasPrevPage       bool
		HasNextPage       bool
		StatusFilter      string
		ServiceTypeFilter string
		SuccessMsg        string
		PageTitle         string
		CustomCSS         string
		CustomJS          string
		AdditionalScripts []string
		CurrentYear       int
	}{
		UserName:          userName,
		Requests:          requests,
		ServiceTypes:      serviceTypes,
		Statuses:          statuses,
		PendingCount:      c.countByStatus(allRequests, 1),
		ConfirmedCount:    c.countByStatus(allRequests, 2),
		CompletedCount:    c.countByStatus(allRequests, 3),
		CanceledCount:     c.countByStatus(allRequests, 4),
		CurrentPage:       page,
		TotalPages:        totalPages,
		TotalCount:        totalCount,
		HasPrevPage:       page > 1,
		HasNextPage:       page < totalPages,
		StatusFilter:      statusFilter,
		ServiceTypeFilter: serviceTypeFilter,
		SuccessMsg:        c.getSuccessMessage(r),
		PageTitle:         "Dashboard Cliente",
		CustomCSS:         "../static/css/cliente_dashboard.css",
		CustomJS:          "../static/js/cliente_dashboard.js",
		AdditionalScripts: []string{},
		CurrentYear:       time.Now().Year(),
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/cliente_dashboard.html",
	}, data)
}

// SolicitarServico - Criar nova solicitação
func (c *ServiceController) SolicitarServico(w http.ResponseWriter, r *http.Request) {
	userID, ok := c.getUserID(w, r)
	if !ok {
		return
	}

	if r.Method == "GET" {
		c.showServiceRequestForm(w, r)
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

	session, _ := config.GetSessionStore().Get(r, "session")
	userName := session.Values["user_name"].(string)

	data := struct {
		Service           *models.ServiceRequest
		UserName          string
		PageTitle         string
		CustomCSS         string
		CustomJS          string
		CurrentYear       int
		AdditionalScripts []string
	}{
		Service:           service,
		UserName:          userName,
		PageTitle:         "Detalhes da Solicitação",
		CustomCSS:         "../static/css/ver_solicitacao.css",
		CustomJS:          "../static/js/ver_solicitacao.js",
		AdditionalScripts: []string{},
		CurrentYear:       time.Now().Year(),
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/ver_solicitacao.html",
	}, data)
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

func (c *ServiceController) renderTemplate(w http.ResponseWriter, templatePaths []string, data interface{}) {
	// Criar template com funções auxiliares
	tmpl := template.New("").Funcs(GetTemplateFuncs())
	
	// Parsear todos os templates
	var err error
	tmpl, err = tmpl.ParseFiles(templatePaths...)
	if err != nil {
		http.Error(w, "Erro ao carregar template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Pegar nome do arquivo base
	baseName := filepath.Base(templatePaths[len(templatePaths)-1])

	// Executar template
	if err := tmpl.ExecuteTemplate(w, baseName, data); err != nil {
		http.Error(w, "Erro ao renderizar template: "+err.Error(), http.StatusInternalServerError)
		return
	}
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

func (c *ServiceController) showServiceRequestForm(w http.ResponseWriter, r *http.Request) {
	serviceTypes, err := c.ServiceModel.GetAllServiceTypes()
	if err != nil {
		http.Error(w, "Erro ao carregar tipos de serviço", http.StatusInternalServerError)
		return
	}

	session, _ := config.GetSessionStore().Get(r, "session")
	userName := session.Values["user_name"].(string)

	data := struct {
		ServiceTypes      []models.ServiceType
		UserName          string
		PageTitle         string
		CustomCSS         string
		CustomJS          string
		CurrentYear       int
		AdditionalScripts []string
	}{
		ServiceTypes:      serviceTypes,
		UserName:          userName,
		PageTitle:         "Solicitar Serviço",
		CustomCSS:         "../static/css/solicitar_servico.css",
		CustomJS:          "../static/js/solicitar_servico.js",
		AdditionalScripts: []string{},
		CurrentYear:       time.Now().Year(),
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/solicitar_servico.html",
	}, data)
}

func (c *ServiceController) showEditForm(w http.ResponseWriter, r *http.Request, requestID, userID int) {
	service, err := c.ServiceModel.GetByIDAndUser(requestID, userID)
	if err != nil {
		c.handleNotFound(w, err, "Solicitação não encontrada")
		return
	}

	if service.StatusID != 1 {
		http.Redirect(w, r, "/solicitacao/"+strconv.Itoa(requestID), http.StatusFound)
		return
	}

	serviceTypes, err := c.ServiceModel.GetAllServiceTypes()
	if err != nil {
		http.Error(w, "Erro ao carregar tipos de serviço", http.StatusInternalServerError)
		return
	}

	session, _ := config.GetSessionStore().Get(r, "session")
	userName, ok := session.Values["user_name"].(string)
	if !ok {
		userName = "Usuário"
	}

	data := struct {
		Service           *models.ServiceRequest
		ServiceTypes      []models.ServiceType
		UserName          string
		PageTitle         string
		CustomCSS         string
		CustomJS          string
		CurrentYear       int
		AdditionalScripts []string
	}{
		Service:           service,
		ServiceTypes:      serviceTypes,
		UserName:          userName,
		PageTitle:         "Editar Solicitação",
		CustomCSS:         "/static/css/solicitar_servico.css",
		CustomJS:          "/static/js/solicitar_servico.js",
		AdditionalScripts: []string{},
		CurrentYear:       time.Now().Year(),
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/editar_solicitacao.html",
	}, data)
}