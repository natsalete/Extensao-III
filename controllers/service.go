package controllers

import (
	"fmt"
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
	userName := session.Values["user_name"].(string)

	data := struct {
		UserName          string
		Requests          []models.ServiceRequest
		PendingCount      int
		ConfirmedCount    int
		CompletedCount    int
		CanceledCount     int
		SuccessMsg        string
		PageTitle         string
		CustomCSS         string
		CustomJS          string
		AdditionalScripts []string 
		CurrentYear       int
	}{
		UserName:          userName,
		Requests:          requests,
		PendingCount:      c.countByStatus(requests, 1),
		ConfirmedCount:    c.countByStatus(requests, 2),
		CompletedCount:    c.countByStatus(requests, 3),
		CanceledCount:     c.countByStatus(requests, 4),
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
		Service     *models.ServiceRequest
		UserName    string
		PageTitle   string
		CustomCSS   string
		CustomJS    string
		CurrentYear int
		AdditionalScripts []string 
	}{
		Service:     service,
		UserName:    userName,
		PageTitle:   "Detalhes da Solicitação",
		CustomCSS:   "../static/css/ver_solicitacao.css",
		CustomJS:    "../static/js/ver_solicitacao.js",
		AdditionalScripts: []string{},  
		CurrentYear: time.Now().Year(),
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
	fmt.Println("\n=== DEBUG renderTemplate ===")
	fmt.Println("Template paths:")
	for i, path := range templatePaths {
		fmt.Printf("  [%d] %s\n", i, path)
	}
	
	tmpl, err := template.ParseFiles(templatePaths...)
	if err != nil {
		fmt.Println(" ERRO ao parsear templates:", err)
		http.Error(w, "Erro ao carregar template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("✅ Templates parseados com sucesso")
	
	// Pega o nome do arquivo base (último da lista)
	baseName := filepath.Base(templatePaths[len(templatePaths)-1])
	fmt.Println("Base template name:", baseName)
	
	// Debug: listar todos os templates parseados
	fmt.Println("Templates disponíveis:")
	for _, t := range tmpl.Templates() {
		fmt.Printf("  - %s\n", t.Name())
	}
	
	fmt.Println("Tentando executar template:", baseName)
	if err := tmpl.ExecuteTemplate(w, baseName, data); err != nil {
		fmt.Println(" ERRO ao executar template:", err)
		http.Error(w, "Erro ao renderizar template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("✅ Template executado com sucesso")
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
		ServiceTypes []models.ServiceType
		UserName     string
		PageTitle    string
		CustomCSS    string
		CustomJS     string
		CurrentYear  int
		AdditionalScripts []string 
	}{
		ServiceTypes: serviceTypes,
		UserName:     userName,
		PageTitle:    "Solicitar Serviço",
		CustomCSS:    "../static/css/solicitar_servico.css",
		CustomJS:     "../static/js/solicitar_servico.js",
		AdditionalScripts: []string{},  
		CurrentYear:  time.Now().Year(),
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
	fmt.Println("=== DEBUG showEditForm ===")
	fmt.Println("RequestID:", requestID, "UserID:", userID)
	
	service, err := c.ServiceModel.GetByIDAndUser(requestID, userID)
	if err != nil {
		fmt.Println("Erro ao buscar service:", err)
		c.handleNotFound(w, err, "Solicitação não encontrada")
		return
	}
	fmt.Println("Service encontrado:", service.ID)

	if service.StatusID != 1 {
		fmt.Println("Status não é 1, redirecionando. StatusID:", service.StatusID)
		http.Redirect(w, r, "/solicitacao/"+strconv.Itoa(requestID), http.StatusFound)
		return
	}

	serviceTypes, err := c.ServiceModel.GetAllServiceTypes()
	if err != nil {
		fmt.Println("Erro ao buscar service types:", err)
		http.Error(w, "Erro ao carregar tipos de serviço", http.StatusInternalServerError)
		return
	}
	fmt.Println("Service types carregados:", len(serviceTypes))

	session, _ := config.GetSessionStore().Get(r, "session")
	userName, ok := session.Values["user_name"].(string)
	if !ok {
		fmt.Println("Erro: user_name não encontrado na sessão")
		userName = "Usuário"
	}
	fmt.Println("UserName:", userName)

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

	fmt.Println("Chamando renderTemplate...")
	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/editar_solicitacao.html",
	}, data)
	fmt.Println("renderTemplate concluído")
}