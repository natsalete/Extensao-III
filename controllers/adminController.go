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
	"martins-pocos/services"

	"github.com/gorilla/mux"
)

type AdminController struct {
	ServiceModel    *models.ServiceModel
	WhatsAppService *services.WhatsAppService
}

func NewAdminController(serviceModel *models.ServiceModel, whatsappService *services.WhatsAppService) *AdminController {
	return &AdminController{
		ServiceModel:    serviceModel,
		WhatsAppService: whatsappService,
	}
}

// AdminDashboard - Dashboard com paginação (5 por página)
func (c *AdminController) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	// Capturar parâmetros de filtro e paginação
	statusFilter := r.URL.Query().Get("status")
	serviceTypeFilter := r.URL.Query().Get("service_type")
	searchQuery := r.URL.Query().Get("search")
	pageStr := r.URL.Query().Get("page")

	// Configurar paginação
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 5 // MÁXIMO 5 POR PÁGINA
	offset := (page - 1) * pageSize

	// Buscar solicitações com filtros e paginação
	requests, totalCount, err := c.ServiceModel.GetAllWithFilters(
		statusFilter,
		serviceTypeFilter,
		searchQuery,
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

	// Buscar estatísticas por status (para os cards)
	stats, err := c.ServiceModel.GetStatusStats()
	if err != nil {
		stats = make(map[string]int)
	}

	session, _ := config.GetSessionStore().Get(r, "session")
	userName := session.Values["user_name"].(string)

	data := struct {
		UserName          string
		Requests          []models.ServiceRequest
		ServiceTypes      []models.ServiceType
		Statuses          []models.RequestStatus
		SolicitadaCount   int
		ConfirmadaCount   int
		RealizadaCount    int
		CanceladaCount    int
		CurrentPage       int
		TotalPages        int
		TotalCount        int
		HasPrevPage       bool
		HasNextPage       bool
		StatusFilter      string
		ServiceTypeFilter string
		SearchQuery       string
		SuccessMsg        string
		PageTitle         string
		CustomCSS         string
		CustomJS          string
		AdditionalScripts []string
		CurrentYear       int
		IsAdmin		  bool
	}{
		UserName:          userName,
		Requests:          requests,
		ServiceTypes:      serviceTypes,
		Statuses:          statuses,
		SolicitadaCount:   stats["solicitada"],
		ConfirmadaCount:   stats["confirmada"],
		RealizadaCount:    stats["realizada"],
		CanceladaCount:    stats["cancelada"],
		CurrentPage:       page,
		TotalPages:        totalPages,
		TotalCount:        totalCount,
		HasPrevPage:       page > 1,
		HasNextPage:       page < totalPages,
		StatusFilter:      statusFilter,
		ServiceTypeFilter: serviceTypeFilter,
		SearchQuery:       searchQuery,
		SuccessMsg:        c.getSuccessMessage(r),
		PageTitle:         "Dashboard Administrativo",
		CustomCSS:         "/static/css/admin.css",
		CustomJS:          "/static/js/admin.js",
		AdditionalScripts: []string{},
		CurrentYear:       time.Now().Year(),
		IsAdmin:		  true,
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/admin_dashboard.html",
	}, data)
}

// VerSolicitacaoAdmin - Ver detalhes da solicitação (admin)
func (c *AdminController) VerSolicitacaoAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	service, err := c.ServiceModel.GetByID(requestID)
	if err != nil {
		http.Error(w, "Solicitação não encontrada", http.StatusNotFound)
		return
	}

	// Buscar todos os status para o modal de edição
	statuses, err := c.ServiceModel.GetAllStatuses()
	if err != nil {
		http.Error(w, "Erro ao carregar status", http.StatusInternalServerError)
		return
	}

	session, _ := config.GetSessionStore().Get(r, "session")
	userName := session.Values["user_name"].(string)

	data := struct {
		Service           *models.ServiceRequest
		Statuses          []models.RequestStatus
		UserName          string
		PageTitle         string
		CustomCSS         string
		CustomJS          string
		CurrentYear       int
		AdditionalScripts []string
		IsAdmin		  bool
	}{
		Service:           service,
		Statuses:          statuses,
		UserName:          userName,
		PageTitle:         "Detalhes da Solicitação",
		CustomCSS:         "/static/css/admin.css",
		CustomJS:          "/static/js/admin.js",
		AdditionalScripts: []string{},
		CurrentYear:       time.Now().Year(),
		IsAdmin:		  true,
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/admin_ver_solicitacao.html",
	}, data)
}

// EditarSolicitacaoAdmin - Editar solicitação (admin pode editar tudo)
func (c *AdminController) EditarSolicitacaoAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		c.showAdminEditForm(w, r, requestID)
		return
	}

	c.updateAdminServiceRequest(w, r, requestID)
}

// UpdateStatus - Atualizar status da solicitação
func (c *AdminController) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestIDStr := r.FormValue("request_id")
	statusIDStr := r.FormValue("status_id")

	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		http.Error(w, "ID da solicitação inválido", http.StatusBadRequest)
		return
	}

	statusID, err := strconv.Atoi(statusIDStr)
	if err != nil {
		http.Error(w, "Status inválido", http.StatusBadRequest)
		return
	}

	// Buscar a solicitação atual
	service, err := c.ServiceModel.GetByID(requestID)
	if err != nil {
		http.Error(w, "Solicitação não encontrada", http.StatusNotFound)
		return
	}

	// Verificar se o status realmente mudou
	if service.StatusID == statusID {
		http.Redirect(w, r, "/dashboard/admin?success=no_change", http.StatusFound)
		return
	}

	// Atualizar o status
	if err := c.ServiceModel.UpdateStatusByID(requestID, statusID); err != nil {
		http.Error(w, "Erro ao atualizar status", http.StatusInternalServerError)
		return
	}

	// Buscar informações do usuário para enviar WhatsApp
	user, err := c.getUserByID(service.UserID)
	if err == nil && user.Phone != "" {
		// Enviar mensagem do WhatsApp
		c.sendWhatsAppNotification(service, user, statusID)
	}

	http.Redirect(w, r, "/dashboard/admin?success=status_updated", http.StatusFound)
}

// DeletarSolicitacao - Deletar solicitação
func (c *AdminController) DeletarSolicitacao(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := c.ServiceModel.Delete(requestID); err != nil {
		http.Error(w, "Erro ao deletar solicitação", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard/admin?success=deleted", http.StatusFound)
}

// Helper methods

func (c *AdminController) getSuccessMessage(r *http.Request) string {
	switch r.URL.Query().Get("success") {
	case "status_updated":
		return "Status atualizado com sucesso!"
	case "updated":
		return "Solicitação atualizada com sucesso!"
	case "deleted":
		return "Solicitação deletada com sucesso!"
	case "no_change":
		return "Nenhuma alteração foi feita"
	default:
		return ""
	}
}

func (c *AdminController) renderTemplate(w http.ResponseWriter, templatePaths []string, data interface{}) {
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

func (c *AdminController) getUserByID(userID int) (*models.User, error) {
	userModel := models.NewUserModel(config.GetDB())
	return userModel.GetByID(userID)
}

func (c *AdminController) sendWhatsAppNotification(service *models.ServiceRequest, user *models.User, newStatusID int) {
	var message string

	switch newStatusID {
	case 2: // CONFIRMADA
		message = fmt.Sprintf(
			"🔔 *Vistoria Confirmada*\n\nOlá %s! ✅\n\nSua vistoria foi confirmada para o dia %s às %s.\n\n📍 *Local:* %s, %s\n%s - %s/%s\n\nEm caso de dúvidas, entre em contato conosco!\n\n_Martins Poços - Sistema Automatizado_",
			service.FullName,
			service.PreferredDate.Format("02/01/2006"),
			service.PreferredTime[:5],
			service.Logradouro,
			service.Numero,
			service.Bairro,
			service.Cidade,
			service.Estado,
		)
	case 3: // REALIZADA
		message = fmt.Sprintf(
			"✅ *Vistoria Realizada*\n\nOlá %s! 🎉\n\nSua vistoria foi realizada e aprovada!\n\nEm breve entraremos em contato para elaboração do contrato.\n\nAcompanhe o status no nosso sistema.\n\n_Martins Poços - Sistema Automatizado_",
			service.FullName,
		)
	case 4: // CANCELADA
		message = fmt.Sprintf(
			"❌ *Vistoria Cancelada*\n\nOlá %s!\n\nInfelizmente sua vistoria foi cancelada.\n\nPara reagendar, acesse nosso sistema ou entre em contato.\n\n_Martins Poços - Sistema Automatizado_",
			service.FullName,
		)
	default:
		return
	}

	// Enviar mensagem
	c.WhatsAppService.SendMessage(user.Phone, message)
}

func (c *AdminController) showAdminEditForm(w http.ResponseWriter, r *http.Request, requestID int) {
	service, err := c.ServiceModel.GetByID(requestID)
	if err != nil {
		http.Error(w, "Solicitação não encontrada", http.StatusNotFound)
		return
	}

	serviceTypes, err := c.ServiceModel.GetAllServiceTypes()
	if err != nil {
		http.Error(w, "Erro ao carregar tipos de serviço", http.StatusInternalServerError)
		return
	}

	statuses, err := c.ServiceModel.GetAllStatuses()
	if err != nil {
		http.Error(w, "Erro ao carregar status", http.StatusInternalServerError)
		return
	}

	session, _ := config.GetSessionStore().Get(r, "session")
	userName := session.Values["user_name"].(string)

	data := struct {
		Service           *models.ServiceRequest
		ServiceTypes      []models.ServiceType
		Statuses          []models.RequestStatus
		UserName          string
		PageTitle         string
		CustomCSS         string
		CustomJS          string
		CurrentYear       int
		AdditionalScripts []string
		IsAdmin		  bool
	}{
		Service:           service,
		ServiceTypes:      serviceTypes,
		Statuses:          statuses,
		UserName:          userName,
		PageTitle:         "Editar Solicitação",
		CustomCSS:         "/static/css/admin.css",
		CustomJS:          "/static/js/solicitar_servico.js",
		AdditionalScripts: []string{},
		CurrentYear:       time.Now().Year(),
		IsAdmin:		  true,
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/admin_editar_solicitacao.html",
	}, data)
}

func (c *AdminController) updateAdminServiceRequest(w http.ResponseWriter, r *http.Request, requestID int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	serviceTypeCode := r.FormValue("service_type")
	serviceType, err := c.ServiceModel.GetServiceTypeByCode(serviceTypeCode)
	if err != nil {
		http.Error(w, "Tipo de serviço inválido", http.StatusBadRequest)
		return
	}

	statusIDStr := r.FormValue("status_id")
	statusID, err := strconv.Atoi(statusIDStr)
	if err != nil {
		http.Error(w, "Status inválido", http.StatusBadRequest)
		return
	}

	preferredDate, err := time.Parse("2006-01-02", r.FormValue("preferred_date"))
	if err != nil {
		http.Error(w, "Data inválida", http.StatusBadRequest)
		return
	}

	service := &models.ServiceRequest{
		ID:            requestID,
		FullName:      r.FormValue("full_name"),
		ServiceTypeID: serviceType.ID,
		Description:   r.FormValue("description"),
		CEP:           r.FormValue("cep"),
		Logradouro:    r.FormValue("logradouro"),
		Numero:        r.FormValue("numero"),
		Bairro:        r.FormValue("bairro"),
		Cidade:        r.FormValue("cidade"),
		Estado:        r.FormValue("estado"),
		PreferredDate: preferredDate,
		PreferredTime: r.FormValue("preferred_time"),
		StatusID:      statusID,
	}

	if err := c.ServiceModel.AdminUpdate(service); err != nil {
		http.Error(w, "Erro ao atualizar solicitação", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard/admin?success=updated", http.StatusFound)
}