package models

import (
	"log"
	"database/sql"
	"fmt"
	"time"
)

// Tipos de garantia disponíveis
const (
	GuaranteeSecondAttempt = "SEGUNDA_TENTATIVA"
	GuaranteeNone          = "SEM_GARANTIA"
	GuaranteeCustom        = "PERSONALIZADA"
)

// Status do contrato
const (
	ContractStatusDraft     = "RASCUNHO"
	ContractStatusPending   = "AGUARDANDO_ASSINATURAS"
	ContractStatusSigned    = "ASSINADO"
	ContractStatusCancelled = "CANCELADO"
)

type Contract struct {
	ID                 int            `json:"id"`
	ServiceRequestID   int            `json:"service_request_id"`
	ContractNumber     string         `json:"contract_number"`
	TotalValue         float64        `json:"total_value"`
	PaymentConditions  string         `json:"payment_conditions"`
	GuaranteeType      string         `json:"guarantee_type"`
	GuaranteeCustom    sql.NullString `json:"guarantee_custom"`
	ClientRequirements sql.NullString `json:"client_requirements"`
	MaterialsUsed      sql.NullString `json:"materials_used"`
	AdditionalNotes    sql.NullString `json:"additional_notes"`
	ClientSigned       bool           `json:"client_signed"`
	ClientSignedAt     sql.NullTime   `json:"client_signed_at"`
	ClientSignature    sql.NullString `json:"client_signature"`
	CompanySigned      bool           `json:"company_signed"`
	CompanySignedAt    sql.NullTime   `json:"company_signed_at"`
	CompanySignature   sql.NullString `json:"company_signature"`
	Status             string         `json:"status"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	// Campos relacionados
	ServiceRequest *ServiceRequest `json:"service_request,omitempty"`
}

type ContractHistory struct {
	ID            int       `json:"id"`
	ContractID    int       `json:"contract_id"`
	Action        string    `json:"action"`
	ChangedBy     int       `json:"changed_by"`
	ChangedFields string    `json:"changed_fields"`
	CreatedAt     time.Time `json:"created_at"`
}

type ContractModel struct {
	DB *sql.DB
}

func NewContractModel(db *sql.DB) *ContractModel {
	return &ContractModel{DB: db}
}

// GenerateContractNumber gera um número único para o contrato
func (m *ContractModel) GenerateContractNumber() string {
	year := time.Now().Year()
	var count int
	m.DB.QueryRow("SELECT COUNT(*) FROM contracts WHERE EXTRACT(YEAR FROM created_at) = $1", year).Scan(&count)
	return fmt.Sprintf("MP-%d-%04d", year, count+1)
}

// Create cria um novo contrato
func (m *ContractModel) Create(contract *Contract) error {
	contract.ContractNumber = m.GenerateContractNumber()
	contract.Status = ContractStatusDraft

	query := `
		INSERT INTO contracts (
			service_request_id, contract_number, total_value, payment_conditions,
			guarantee_type, guarantee_custom, client_requirements, materials_used,
			additional_notes, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	return m.DB.QueryRow(
		query,
		contract.ServiceRequestID, contract.ContractNumber, contract.TotalValue,
		contract.PaymentConditions, contract.GuaranteeType, contract.GuaranteeCustom,
		contract.ClientRequirements, contract.MaterialsUsed, contract.AdditionalNotes,
		contract.Status,
	).Scan(&contract.ID, &contract.CreatedAt, &contract.UpdatedAt)
}

// Update atualiza um contrato (apenas se não estiver assinado)
func (m *ContractModel) Update(contract *Contract) error {
	query := `
		UPDATE contracts SET
			total_value = $1, payment_conditions = $2, guarantee_type = $3,
			guarantee_custom = $4, client_requirements = $5, materials_used = $6,
			additional_notes = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8 AND status IN ('RASCUNHO', 'AGUARDANDO_ASSINATURAS') 
		AND client_signed = false AND company_signed = false`

	result, err := m.DB.Exec(
		query,
		contract.TotalValue, contract.PaymentConditions, contract.GuaranteeType,
		contract.GuaranteeCustom, contract.ClientRequirements, contract.MaterialsUsed,
		contract.AdditionalNotes, contract.ID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("contrato não pode ser editado (já assinado ou não encontrado)")
	}
	return nil
}

// SendForSignature envia o contrato para assinatura
func (m *ContractModel) SendForSignature(contractID int) error {
	log.Printf("🔵 [MODEL] SendForSignature iniciado para contrato ID: %d", contractID)
	
	// Verificar se o contrato existe e está em rascunho
	var currentStatus string
	checkQuery := `SELECT status FROM contracts WHERE id = $1`
	
	log.Printf("🔍 [MODEL] Executando query: %s com ID: %d", checkQuery, contractID)
	err := m.DB.QueryRow(checkQuery, contractID).Scan(&currentStatus)
	
	if err == sql.ErrNoRows {
		log.Printf("❌ [MODEL] Contrato #%d NÃO ENCONTRADO no banco", contractID)
		return fmt.Errorf("contrato #%d não encontrado", contractID)
	}
	if err != nil {
		log.Printf("❌ [MODEL] ERRO ao buscar contrato: %v", err)
		return fmt.Errorf("erro ao buscar contrato: %w", err)
	}
	
	log.Printf("✅ [MODEL] Contrato encontrado. Status atual: '%s'", currentStatus)
	
	if currentStatus != "RASCUNHO" {
		log.Printf("⚠️ [MODEL] Status incorreto! Esperado: RASCUNHO, Recebido: %s", currentStatus)
		return fmt.Errorf("contrato não pode ser enviado - status atual: %s (deve estar em RASCUNHO)", currentStatus)
	}
	
	// Atualizar status
	updateQuery := `UPDATE contracts 
	                SET status = $1, updated_at = CURRENT_TIMESTAMP 
	                WHERE id = $2`
	
	newStatus := "AGUARDANDO_ASSINATURAS"
	log.Printf("📝 [MODEL] Atualizando status de '%s' para '%s'", currentStatus, newStatus)
	log.Printf("🔍 [MODEL] Query: %s", updateQuery)
	log.Printf("🔍 [MODEL] Parâmetros: status='%s', id=%d", newStatus, contractID)
	
	result, err := m.DB.Exec(updateQuery, newStatus, contractID)
	if err != nil {
		log.Printf("❌ [MODEL] ERRO ao executar UPDATE: %v", err)
		return fmt.Errorf("erro ao atualizar status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("⚠️ [MODEL] Erro ao verificar linhas afetadas: %v", err)
	} else {
		log.Printf("✅ [MODEL] Linhas afetadas: %d", rowsAffected)
	}
	
	log.Printf("✅ [MODEL] SendForSignature concluído com sucesso!")
	return nil
}

// SignByClient assina o contrato pelo cliente - COM VERIFICAÇÕES
func (m *ContractModel) SignByClient(contractID int, signature string) error {
	// Verificar se o contrato existe e está aguardando assinaturas
	var status string
	var clientSigned bool
	checkQuery := `SELECT status, client_signed FROM contracts WHERE id = $1`
	err := m.DB.QueryRow(checkQuery, contractID).Scan(&status, &clientSigned)
	
	if err == sql.ErrNoRows {
		return fmt.Errorf("contrato não encontrado")
	}
	if err != nil {
		return fmt.Errorf("erro ao verificar contrato: %w", err)
	}
	
	if status != "AGUARDANDO_ASSINATURAS" {
		return fmt.Errorf("contrato não está aguardando assinaturas (status: %s)", status)
	}
	
	if clientSigned {
		return fmt.Errorf("contrato já foi assinado pelo cliente")
	}
	
	// Iniciar transação
	tx, err := m.DB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Assinar
	query := `UPDATE contracts 
	          SET client_signed = true, 
	              client_signed_at = CURRENT_TIMESTAMP,
	              client_signature = $1, 
	              updated_at = CURRENT_TIMESTAMP
	          WHERE id = $2`

	_, err = tx.Exec(query, signature, contractID)
	if err != nil {
		return fmt.Errorf("erro ao assinar: %w", err)
	}

	// Verificar se ambos assinaram e finalizar contrato
	m.checkAndFinalizeContract(tx, contractID)
	
	return tx.Commit()
}

// SignByCompany assina o contrato pela empresa - COM VERIFICAÇÕES
func (m *ContractModel) SignByCompany(contractID int, signature string) error {
	// Verificar se o contrato existe e está aguardando assinaturas
	var status string
	var companySigned bool
	checkQuery := `SELECT status, company_signed FROM contracts WHERE id = $1`
	err := m.DB.QueryRow(checkQuery, contractID).Scan(&status, &companySigned)
	
	if err == sql.ErrNoRows {
		return fmt.Errorf("contrato não encontrado")
	}
	if err != nil {
		return fmt.Errorf("erro ao verificar contrato: %w", err)
	}
	
	if status != "AGUARDANDO_ASSINATURAS" {
		return fmt.Errorf("contrato não está aguardando assinaturas (status: %s)", status)
	}
	
	if companySigned {
		return fmt.Errorf("contrato já foi assinado pela empresa")
	}
	
	// Iniciar transação
	tx, err := m.DB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Assinar
	query := `UPDATE contracts 
	          SET company_signed = true, 
	              company_signed_at = CURRENT_TIMESTAMP,
	              company_signature = $1, 
	              updated_at = CURRENT_TIMESTAMP
	          WHERE id = $2`

	_, err = tx.Exec(query, signature, contractID)
	if err != nil {
		return fmt.Errorf("erro ao assinar: %w", err)
	}

	// Verificar se ambos assinaram e finalizar
	m.checkAndFinalizeContract(tx, contractID)
	
	return tx.Commit()
}

// checkAndFinalizeContract verifica se ambos assinaram e finaliza
func (m *ContractModel) checkAndFinalizeContract(tx *sql.Tx, contractID int) {
	var clientSigned, companySigned bool
	tx.QueryRow("SELECT client_signed, company_signed FROM contracts WHERE id = $1", contractID).
		Scan(&clientSigned, &companySigned)

	if clientSigned && companySigned {
		tx.Exec("UPDATE contracts SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", 
			"ASSINADO", contractID)
	}
}

// GetByID busca um contrato pelo ID
func (m *ContractModel) GetByID(id int) (*Contract, error) {
	contract := &Contract{}
	query := `SELECT id, service_request_id, contract_number, total_value, payment_conditions,
		guarantee_type, guarantee_custom, client_requirements, materials_used,
		additional_notes, client_signed, client_signed_at, client_signature,
		company_signed, company_signed_at, company_signature, status, created_at, updated_at
		FROM contracts WHERE id = $1`

	err := m.DB.QueryRow(query, id).Scan(
		&contract.ID, &contract.ServiceRequestID, &contract.ContractNumber,
		&contract.TotalValue, &contract.PaymentConditions, &contract.GuaranteeType,
		&contract.GuaranteeCustom, &contract.ClientRequirements, &contract.MaterialsUsed,
		&contract.AdditionalNotes, &contract.ClientSigned, &contract.ClientSignedAt,
		&contract.ClientSignature, &contract.CompanySigned, &contract.CompanySignedAt,
		&contract.CompanySignature, &contract.Status, &contract.CreatedAt, &contract.UpdatedAt,
	)
	return contract, err
}

// GetByServiceRequestID busca contrato pela solicitação
func (m *ContractModel) GetByServiceRequestID(serviceRequestID int) (*Contract, error) {
	contract := &Contract{}
	query := `SELECT id, service_request_id, contract_number, total_value, payment_conditions,
		guarantee_type, guarantee_custom, client_requirements, materials_used,
		additional_notes, client_signed, client_signed_at, client_signature,
		company_signed, company_signed_at, company_signature, status, created_at, updated_at
		FROM contracts WHERE service_request_id = $1`

	err := m.DB.QueryRow(query, serviceRequestID).Scan(
		&contract.ID, &contract.ServiceRequestID, &contract.ContractNumber,
		&contract.TotalValue, &contract.PaymentConditions, &contract.GuaranteeType,
		&contract.GuaranteeCustom, &contract.ClientRequirements, &contract.MaterialsUsed,
		&contract.AdditionalNotes, &contract.ClientSigned, &contract.ClientSignedAt,
		&contract.ClientSignature, &contract.CompanySigned, &contract.CompanySignedAt,
		&contract.CompanySignature, &contract.Status, &contract.CreatedAt, &contract.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return contract, err
}

// GetAllWithDetails busca todos os contratos com detalhes
func (m *ContractModel) GetAllWithDetails(statusFilter string, limit, offset int) ([]Contract, int, error) {
	baseQuery := ` FROM contracts c JOIN service_requests sr ON c.service_request_id = sr.id WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if statusFilter != "" {
		baseQuery += fmt.Sprintf(" AND c.status = $%d", argPos)
		args = append(args, statusFilter)
		argPos++
	}

	var total int
	countQuery := "SELECT COUNT(*)" + baseQuery
	m.DB.QueryRow(countQuery, args...).Scan(&total)

	selectQuery := fmt.Sprintf(`
		SELECT c.id, c.service_request_id, c.contract_number, c.total_value, c.payment_conditions,
			c.guarantee_type, c.guarantee_custom, c.client_requirements, c.materials_used,
			c.additional_notes, c.client_signed, c.client_signed_at, c.client_signature,
			c.company_signed, c.company_signed_at, c.company_signature, c.status, c.created_at, c.updated_at
		%s ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d`, baseQuery, argPos, argPos+1)

	args = append(args, limit, offset)
	rows, err := m.DB.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contracts []Contract
	for rows.Next() {
		var c Contract
		err := rows.Scan(
			&c.ID, &c.ServiceRequestID, &c.ContractNumber, &c.TotalValue, &c.PaymentConditions,
			&c.GuaranteeType, &c.GuaranteeCustom, &c.ClientRequirements, &c.MaterialsUsed,
			&c.AdditionalNotes, &c.ClientSigned, &c.ClientSignedAt, &c.ClientSignature,
			&c.CompanySigned, &c.CompanySignedAt, &c.CompanySignature, &c.Status, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		contracts = append(contracts, c)
	}
	return contracts, total, nil
}

// GetPendingForClient busca contratos pendentes de assinatura do cliente
func (m *ContractModel) GetPendingForClient(userID int) ([]Contract, error) {
	query := `
		SELECT c.id, c.service_request_id, c.contract_number, c.total_value, c.status, c.created_at
		FROM contracts c
		JOIN service_requests sr ON c.service_request_id = sr.id
		WHERE sr.user_id = $1 AND c.status = 'AGUARDANDO_ASSINATURAS' AND c.client_signed = false`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []Contract
	for rows.Next() {
		var c Contract
		rows.Scan(&c.ID, &c.ServiceRequestID, &c.ContractNumber, &c.TotalValue, &c.Status, &c.CreatedAt)
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// CanEdit verifica se o contrato pode ser editado
func (m *ContractModel) CanEdit(contractID int) bool {
	var status string
	var clientSigned, companySigned bool
	m.DB.QueryRow("SELECT status, client_signed, company_signed FROM contracts WHERE id = $1", contractID).Scan(&status, &clientSigned, &companySigned)
	return (status == ContractStatusDraft || status == ContractStatusPending) && !clientSigned && !companySigned
}

// AddHistory adiciona registro no histórico
func (m *ContractModel) AddHistory(contractID, userID int, action, fields string) error {
	query := `INSERT INTO contract_history (contract_id, action, changed_by, changed_fields) VALUES ($1, $2, $3, $4)`
	_, err := m.DB.Exec(query, contractID, action, userID, fields)
	return err
}

// GetAllByUserID busca todos os contratos de um usuário específico com filtros
func (m *ContractModel) GetAllByUserID(userID int, statusFilter string, limit, offset int) ([]Contract, int, error) {
	baseQuery := ` FROM contracts c 
	               JOIN service_requests sr ON c.service_request_id = sr.id 
	               WHERE sr.user_id = $1`
	args := []interface{}{userID}
	argPos := 2

	if statusFilter != "" {
		baseQuery += fmt.Sprintf(" AND c.status = $%d", argPos)
		args = append(args, statusFilter)
		argPos++
	}

	// Contar total
	var total int
	countQuery := "SELECT COUNT(*)" + baseQuery
	err := m.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Buscar registros
	selectQuery := fmt.Sprintf(`
		SELECT c.id, c.service_request_id, c.contract_number, c.total_value, c.payment_conditions,
			c.guarantee_type, c.guarantee_custom, c.client_requirements, c.materials_used,
			c.additional_notes, c.client_signed, c.client_signed_at, c.client_signature,
			c.company_signed, c.company_signed_at, c.company_signature, c.status, c.created_at, c.updated_at
		%s ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d`, baseQuery, argPos, argPos+1)

	args = append(args, limit, offset)
	rows, err := m.DB.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contracts []Contract
	for rows.Next() {
		var c Contract
		err := rows.Scan(
			&c.ID, &c.ServiceRequestID, &c.ContractNumber, &c.TotalValue, &c.PaymentConditions,
			&c.GuaranteeType, &c.GuaranteeCustom, &c.ClientRequirements, &c.MaterialsUsed,
			&c.AdditionalNotes, &c.ClientSigned, &c.ClientSignedAt, &c.ClientSignature,
			&c.CompanySigned, &c.CompanySignedAt, &c.CompanySignature, &c.Status, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		contracts = append(contracts, c)
	}
	
	return contracts, total, nil
}