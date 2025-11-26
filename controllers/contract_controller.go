package controllers

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"martins-pocos/config"
	"martins-pocos/models"

	"github.com/gorilla/mux"
)

type ContractController struct {
	ContractModel *models.ContractModel
	ServiceModel  *models.ServiceModel
}

func NewContractController(contractModel *models.ContractModel, serviceModel *models.ServiceModel) *ContractController {
	return &ContractController{
		ContractModel: contractModel,
		ServiceModel:  serviceModel,
	}
}

// ListContracts - Lista todos os contratos (admin)
func (c *ContractController) ListContracts(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 10
	offset := (page - 1) * pageSize

	contracts, totalCount, err := c.ContractModel.GetAllWithDetails(statusFilter, pageSize, offset)
	if err != nil {
		http.Error(w, "Erro ao buscar contratos", http.StatusInternalServerError)
		return
	}

	// Enriquecer com dados da solicitação
	for i := range contracts {
		sr, _ := c.ServiceModel.GetByID(contracts[i].ServiceRequestID)
		contracts[i].ServiceRequest = sr
	}

	totalPages := (totalCount + pageSize - 1) / pageSize

	session, _ := config.GetSessionStore().Get(r, "session")
	userName := session.Values["user_name"].(string)

	data := struct {
		Contracts     []models.Contract
		UserName      string
		PageTitle     string
		CustomCSS     string
		CustomJS      string
		CurrentYear   int
		CurrentPage   int
		TotalPages    int
		TotalCount    int
		HasPrevPage   bool
		HasNextPage   bool
		StatusFilter  string
		SuccessMsg    string
		IsAdmin       bool
		AdditionalScripts []string
	}{
		Contracts:     contracts,
		UserName:      userName,
		PageTitle:     "Gestão de Contratos",
		CustomCSS:     "/static/css/contracts.css",
		CustomJS:      "/static/js/contracts.js",
		CurrentYear:   time.Now().Year(),
		CurrentPage:   page,
		TotalPages:    totalPages,
		TotalCount:    totalCount,
		HasPrevPage:   page > 1,
		HasNextPage:   page < totalPages,
		StatusFilter:  statusFilter,
		SuccessMsg:    c.getSuccessMsg(r),
		IsAdmin:       true,
		AdditionalScripts: []string{},
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/admin_contratos.html",
	}, data)
}

// CreateContract - Criar contrato para solicitação realizada
func (c *ContractController) CreateContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceRequestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// Verificar se solicitação existe e está realizada
	service, err := c.ServiceModel.GetByID(serviceRequestID)
	if err != nil || service.StatusID != 3 {
		http.Error(w, "Solicitação não encontrada ou não está realizada", http.StatusBadRequest)
		return
	}

	// Verificar se já existe contrato
	existing, _ := c.ContractModel.GetByServiceRequestID(serviceRequestID)
	if existing != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/contratos/%d", existing.ID), http.StatusFound)
		return
	}

	if r.Method == "GET" {
		c.showCreateForm(w, r, service)
		return
	}

	c.processCreateContract(w, r, serviceRequestID)
}

func (c *ContractController) showCreateForm(w http.ResponseWriter, r *http.Request, service *models.ServiceRequest) {
	session, _ := config.GetSessionStore().Get(r, "session")
	userName := session.Values["user_name"].(string)

	data := struct {
		Service     *models.ServiceRequest
		UserName    string
		PageTitle   string
		CustomCSS   string
		CustomJS    string
		CurrentYear int
		IsAdmin     bool
		AdditionalScripts []string
	}{
		Service:     service,
		UserName:    userName,
		PageTitle:   "Criar Contrato",
		CustomCSS:   "/static/css/contracts.css",
		CustomJS:    "/static/js/contracts.js",
		CurrentYear: time.Now().Year(),
		IsAdmin:     true,
		AdditionalScripts: []string{},
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/admin_criar_contrato.html",
	}, data)
}

func (c *ContractController) processCreateContract(w http.ResponseWriter, r *http.Request, serviceRequestID int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	totalValue, _ := strconv.ParseFloat(r.FormValue("total_value"), 64)

	contract := &models.Contract{
		ServiceRequestID:   serviceRequestID,
		TotalValue:         totalValue,
		PaymentConditions:  r.FormValue("payment_conditions"),
		GuaranteeType:      r.FormValue("guarantee_type"),
		GuaranteeCustom:    toNullString(r.FormValue("guarantee_custom")),
		ClientRequirements: toNullString(r.FormValue("client_requirements")),
		MaterialsUsed:      toNullString(r.FormValue("materials_used")),
		AdditionalNotes:    toNullString(r.FormValue("additional_notes")),
	}

	if err := c.ContractModel.Create(contract); err != nil {
		http.Error(w, "Erro ao criar contrato: "+err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := config.GetSessionStore().Get(r, "session")
	userID := session.Values["user_id"].(int)
	c.ContractModel.AddHistory(contract.ID, userID, "CRIADO", "Contrato criado")

	http.Redirect(w, r, fmt.Sprintf("/admin/contratos/%d?success=created", contract.ID), http.StatusFound)
}

// prepareContractForView processa as assinaturas para exibição segura
func prepareContractForView(contract *models.Contract) map[string]interface{} {
	data := make(map[string]interface{})
	
	// Extrair apenas o base64 das assinaturas (remover o prefixo data:image/png;base64,)
	if contract.CompanySignature.Valid && contract.CompanySignature.String != "" {
		base64Data := extractBase64(contract.CompanySignature.String)
		data["CompanySignatureData"] = base64Data
		log.Printf("🔵 Assinatura empresa processada: %d caracteres", len(base64Data))
	} else {
		data["CompanySignatureData"] = ""
		log.Printf("⚠️ Assinatura empresa não disponível")
	}
	
	if contract.ClientSignature.Valid && contract.ClientSignature.String != "" {
		base64Data := extractBase64(contract.ClientSignature.String)
		data["ClientSignatureData"] = base64Data
		log.Printf("🔵 Assinatura cliente processada: %d caracteres", len(base64Data))
	} else {
		data["ClientSignatureData"] = ""
		log.Printf("⚠️ Assinatura cliente não disponível")
	}
	
	return data
}

// extractBase64 remove o prefixo data:image/png;base64, se existir
func extractBase64(dataURL string) string {
	// Remover prefixo se existir
	if strings.HasPrefix(dataURL, "data:image/png;base64,") {
		return strings.TrimPrefix(dataURL, "data:image/png;base64,")
	}
	// Se já for só base64, retorna como está
	return dataURL
}

// validateBase64 verifica se é um base64 válido
func validateBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}



// ViewContract - Ver detalhes do contrato (ADMIN)
func (c *ContractController) ViewContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	contract, err := c.ContractModel.GetByID(contractID)
	if err != nil {
		http.Error(w, "Contrato não encontrado", http.StatusNotFound)
		return
	}

	service, _ := c.ServiceModel.GetByID(contract.ServiceRequestID)
	contract.ServiceRequest = service

	session, _ := config.GetSessionStore().Get(r, "session")
	userName := session.Values["user_name"].(string)

	// Preparar assinaturas
	signatureData := prepareContractForView(contract)
	
	companySignatureData := ""
	clientSignatureData := ""
	
	if val, ok := signatureData["CompanySignatureData"].(string); ok {
		companySignatureData = val
	}
	if val, ok := signatureData["ClientSignatureData"].(string); ok {
		clientSignatureData = val
	}

	data := struct {
		Contract              *models.Contract
		CompanySignatureData  string
		ClientSignatureData   string
		UserName              string
		PageTitle             string
		CustomCSS             string
		CustomJS              string
		CurrentYear           int
		CanEdit               bool
		SuccessMsg            string
		IsAdmin               bool
		AdditionalScripts     []string
	}{
		Contract:             contract,
		CompanySignatureData: companySignatureData,
		ClientSignatureData:  clientSignatureData,
		UserName:             userName,
		PageTitle:            "Contrato " + contract.ContractNumber,
		CustomCSS:            "/static/css/contracts.css",
		CustomJS:             "/static/js/contracts.js",
		CurrentYear:          time.Now().Year(),
		CanEdit:              c.ContractModel.CanEdit(contractID),
		SuccessMsg:           c.getSuccessMsg(r),
		IsAdmin:              true,
		AdditionalScripts:    []string{},
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/admin_ver_contrato.html",
	}, data)
}


// EditContract - Editar contrato
func (c *ContractController) EditContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if !c.ContractModel.CanEdit(contractID) {
		http.Error(w, "Contrato não pode ser editado", http.StatusForbidden)
		return
	}

	contract, err := c.ContractModel.GetByID(contractID)
	if err != nil {
		http.Error(w, "Contrato não encontrado", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		c.showEditForm(w, r, contract)
		return
	}

	c.processEditContract(w, r, contract)
}

func (c *ContractController) showEditForm(w http.ResponseWriter, r *http.Request, contract *models.Contract) {
	service, _ := c.ServiceModel.GetByID(contract.ServiceRequestID)
	contract.ServiceRequest = service

	session, _ := config.GetSessionStore().Get(r, "session")
	userName := session.Values["user_name"].(string)

	data := struct {
		Contract    *models.Contract
		UserName    string
		PageTitle   string
		CustomCSS   string
		CustomJS    string
		CurrentYear int
		IsAdmin     bool
		AdditionalScripts []string
	}{
		Contract:    contract,
		UserName:    userName,
		PageTitle:   "Editar Contrato",
		CustomCSS:   "/static/css/contracts.css",
		CustomJS:    "/static/js/contracts.js",
		CurrentYear: time.Now().Year(),
		IsAdmin:     true,
		AdditionalScripts: []string{},
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/admin_editar_contrato.html",
	}, data)
}

func (c *ContractController) processEditContract(w http.ResponseWriter, r *http.Request, contract *models.Contract) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	totalValue, _ := strconv.ParseFloat(r.FormValue("total_value"), 64)

	contract.TotalValue = totalValue
	contract.PaymentConditions = r.FormValue("payment_conditions")
	contract.GuaranteeType = r.FormValue("guarantee_type")
	contract.GuaranteeCustom = toNullString(r.FormValue("guarantee_custom"))
	contract.ClientRequirements = toNullString(r.FormValue("client_requirements"))
	contract.MaterialsUsed = toNullString(r.FormValue("materials_used"))
	contract.AdditionalNotes = toNullString(r.FormValue("additional_notes"))

	if err := c.ContractModel.Update(contract); err != nil {
		http.Error(w, "Erro ao atualizar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := config.GetSessionStore().Get(r, "session")
	userID := session.Values["user_id"].(int)
	c.ContractModel.AddHistory(contract.ID, userID, "EDITADO", "Contrato atualizado")

	http.Redirect(w, r, fmt.Sprintf("/admin/contratos/%d?success=updated", contract.ID), http.StatusFound)
}

// SendForSignature - Enviar para assinatura
func (c *ContractController) SendForSignature(w http.ResponseWriter, r *http.Request) {
	log.Println("========== INÍCIO SendForSignature ==========")
	
	vars := mux.Vars(r)
	contractIDStr := vars["id"]
	log.Printf("📌 ID do contrato (string): %s", contractIDStr)
	
	contractID, err := strconv.Atoi(contractIDStr)
	if err != nil {
		log.Printf("❌ ERRO ao converter ID: %v", err)
		http.Error(w, "ID inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("✅ ID convertido: %d", contractID)
	
	// Verificar se o contrato existe antes de enviar
	log.Println("🔍 Buscando contrato...")
	contract, err := c.ContractModel.GetByID(contractID)
	if err != nil {
		log.Printf("❌ ERRO ao buscar contrato: %v", err)
		http.Error(w, "Contrato não encontrado: "+err.Error(), http.StatusNotFound)
		return
	}
	log.Printf("✅ Contrato encontrado: #%d | Status: %s", contract.ID, contract.Status)
	
	// Verificar se está em rascunho
	if contract.Status != "RASCUNHO" {
		log.Printf("⚠️ Contrato não está em RASCUNHO. Status atual: %s", contract.Status)
		http.Error(w, fmt.Sprintf("Contrato não pode ser enviado - status atual: %s", contract.Status), http.StatusBadRequest)
		return
	}
	
	log.Println("📤 Enviando para assinatura...")
	err = c.ContractModel.SendForSignature(contractID)
	if err != nil {
		log.Printf("❌ ERRO no SendForSignature: %v", err)
		http.Error(w, "Erro ao enviar para assinatura: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("✅ Status atualizado com sucesso!")

	// Adicionar histórico
	session, err := config.GetSessionStore().Get(r, "session")
	if err != nil {
		log.Printf("⚠️ Erro ao obter sessão: %v", err)
	}
	
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		log.Println("⚠️ user_id não encontrado na sessão")
		userID = 0
	}
	log.Printf("👤 Usuário: %d", userID)
	
	err = c.ContractModel.AddHistory(contractID, userID, "ENVIADO_ASSINATURA", "Enviado para assinatura")
	if err != nil {
		log.Printf("⚠️ Erro ao adicionar histórico: %v", err)
	} else {
		log.Println("✅ Histórico adicionado")
	}

	redirectURL := fmt.Sprintf("/admin/contratos/%d?success=sent", contractID)
	log.Printf("🔄 Redirecionando para: %s", redirectURL)
	log.Println("========== FIM SendForSignature ==========")
	
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// SignContractCompany - Assinar pela empresa
func (c *ContractController) SignContractCompany(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractID, _ := strconv.Atoi(vars["id"])
	signature := r.FormValue("signature")

	if signature == "" {
		http.Error(w, "Assinatura obrigatória", http.StatusBadRequest)
		return
	}

	if err := c.ContractModel.SignByCompany(contractID, signature); err != nil {
		http.Error(w, "Erro ao assinar", http.StatusInternalServerError)
		return
	}

	session, _ := config.GetSessionStore().Get(r, "session")
	userID := session.Values["user_id"].(int)
	c.ContractModel.AddHistory(contractID, userID, "ASSINADO_EMPRESA", "Assinado pela empresa")

	http.Redirect(w, r, fmt.Sprintf("/admin/contratos/%d?success=signed", contractID), http.StatusFound)
}

// ClientContracts - Lista TODOS os contratos do cliente com filtros e paginação
func (c *ContractController) ClientContracts(w http.ResponseWriter, r *http.Request) {
	session, _ := config.GetSessionStore().Get(r, "session")
	userID := session.Values["user_id"].(int)
	userName := session.Values["user_name"].(string)

	// Capturar filtros
	statusFilter := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 9
	offset := (page - 1) * pageSize

	// Buscar contratos do cliente
	contracts, totalCount, err := c.ContractModel.GetAllByUserID(userID, statusFilter, pageSize, offset)
	if err != nil {
		http.Error(w, "Erro ao buscar contratos", http.StatusInternalServerError)
		return
	}

	// Enriquecer com dados da solicitação
	for i := range contracts {
		sr, _ := c.ServiceModel.GetByID(contracts[i].ServiceRequestID)
		contracts[i].ServiceRequest = sr
	}

	totalPages := (totalCount + pageSize - 1) / pageSize

	data := struct {
		Contracts     []models.Contract
		UserName      string
		PageTitle     string
		CustomCSS     string
		CustomJS      string
		CurrentYear   int
		CurrentPage   int
		TotalPages    int
		TotalCount    int
		HasPrevPage   bool
		HasNextPage   bool
		StatusFilter  string
		SuccessMsg    string
		IsAdmin       bool
		AdditionalScripts []string
	}{
		Contracts:     contracts,
		UserName:      userName,
		PageTitle:     "Meus Contratos",
		CustomCSS:     "/static/css/contracts.css",
		CustomJS:      "/static/js/contracts.js",
		CurrentYear:   time.Now().Year(),
		CurrentPage:   page,
		TotalPages:    totalPages,
		TotalCount:    totalCount,
		HasPrevPage:   page > 1,
		HasNextPage:   page < totalPages,
		StatusFilter:  statusFilter,
		SuccessMsg:    c.getSuccessMsg(r),
		IsAdmin:       false,
		AdditionalScripts: []string{},
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/cliente_contratos.html",
	}, data)
}

// ClientViewContract - Cliente visualiza contrato (CORRIGIDO)
func (c *ContractController) ClientViewContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractID, _ := strconv.Atoi(vars["id"])

	session, _ := config.GetSessionStore().Get(r, "session")
	userID := session.Values["user_id"].(int)
	userName := session.Values["user_name"].(string)

	contract, err := c.ContractModel.GetByID(contractID)
	if err != nil {
		http.Error(w, "Contrato não encontrado", http.StatusNotFound)
		return
	}

	service, _ := c.ServiceModel.GetByID(contract.ServiceRequestID)
	if service.UserID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}
	contract.ServiceRequest = service

	// ============================================
	// CORREÇÃO: Preparar assinaturas corretamente
	// ============================================
	signatureData := prepareContractForView(contract)
	
	// Extrair com segurança
	companySignatureData := ""
	clientSignatureData := ""
	
	if val, ok := signatureData["CompanySignatureData"].(string); ok {
		companySignatureData = val
		log.Printf("✅ Assinatura empresa preparada: %d caracteres", len(companySignatureData))
	} else {
		log.Printf("⚠️ Assinatura empresa não disponível")
	}
	
	if val, ok := signatureData["ClientSignatureData"].(string); ok {
		clientSignatureData = val
		log.Printf("✅ Assinatura cliente preparada: %d caracteres", len(clientSignatureData))
	} else {
		log.Printf("⚠️ Assinatura cliente não disponível")
	}

	data := struct {
		Contract              *models.Contract
		CompanySignatureData  string
		ClientSignatureData   string
		UserName              string
		PageTitle             string
		CustomCSS             string
		CustomJS              string
		CurrentYear           int
		SuccessMsg            string
		IsAdmin               bool
		AdditionalScripts     []string
	}{
		Contract:             contract,
		CompanySignatureData: companySignatureData,
		ClientSignatureData:  clientSignatureData, // ← CORRIGIDO
		UserName:             userName,
		PageTitle:            "Contrato " + contract.ContractNumber,
		CustomCSS:            "/static/css/contracts.css",
		CustomJS:             "/static/js/contracts.js",
		CurrentYear:          time.Now().Year(),
		SuccessMsg:           c.getSuccessMsg(r),
		IsAdmin:              false,
		AdditionalScripts:    []string{},
	}

	c.renderTemplate(w, []string{
		"templates/components/head.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/scripts.html",
		"templates/cliente_ver_contrato.html",
	}, data)
}

func (c *ContractController) ClientSignContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractID, _ := strconv.Atoi(vars["id"])

	session, _ := config.GetSessionStore().Get(r, "session")
	userID := session.Values["user_id"].(int)

	contract, _ := c.ContractModel.GetByID(contractID)
	service, _ := c.ServiceModel.GetByID(contract.ServiceRequestID)
	if service.UserID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	signature := r.FormValue("signature")
	
	// LOGS DE DEBUG
	log.Printf("🔵 Assinatura do cliente recebida")
	log.Printf("📏 Tamanho: %d caracteres", len(signature))
	log.Printf("📝 Prefixo: %s", signature[:min(50, len(signature))])
	
	if signature == "" {
		http.Error(w, "Assinatura obrigatória", http.StatusBadRequest)
		return
	}

	// Validar se é base64 válido
	base64Data := extractBase64(signature)
	if !validateBase64(base64Data) {
		log.Printf("❌ Base64 inválido!")
		http.Error(w, "Assinatura inválida", http.StatusBadRequest)
		return
	}

	if err := c.ContractModel.SignByClient(contractID, signature); err != nil {
		log.Printf("❌ Erro ao salvar: %v", err)
		http.Error(w, "Erro ao assinar", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Assinatura salva com sucesso!")

	c.ContractModel.AddHistory(contractID, userID, "ASSINADO_CLIENTE", "Assinado pelo cliente")

	http.Redirect(w, r, fmt.Sprintf("/contratos/%d?success=signed", contractID), http.StatusFound)
}

// Helper min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helpers
func (c *ContractController) getSuccessMsg(r *http.Request) string {
	switch r.URL.Query().Get("success") {
	case "created":
		return "Contrato criado com sucesso!"
	case "updated":
		return "Contrato atualizado com sucesso!"
	case "sent":
		return "Contrato enviado para assinatura!"
	case "signed":
		return "Contrato assinado com sucesso!"
	default:
		return ""
	}
}

func (c *ContractController) renderTemplate(w http.ResponseWriter, paths []string, data interface{}) {
	tmpl := template.New("").Funcs(GetTemplateFuncs())
	tmpl, err := tmpl.ParseFiles(paths...)
	if err != nil {
		http.Error(w, "Erro ao carregar template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	baseName := filepath.Base(paths[len(paths)-1])
	if err := tmpl.ExecuteTemplate(w, baseName, data); err != nil {
		http.Error(w, "Erro ao renderizar: "+err.Error(), http.StatusInternalServerError)
	}
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

