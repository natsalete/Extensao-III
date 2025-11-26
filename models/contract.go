package models

import (
	"log"
	"database/sql"
	"fmt"
	"time"
)

// ============================================
// STRUCTS REFATORADOS
// ============================================

type GuaranteeType struct {
	ID                 int       `json:"id"`
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	RequiresCustomText bool      `json:"requires_custom_text"`
	DisplayOrder       int       `json:"display_order"`
	Active             bool      `json:"active"`
	CreatedAt          time.Time `json:"created_at"`
}

type ContractStatus struct {
	ID           int       `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ColorClass   string    `json:"color_class"`
	BadgeClass   string    `json:"badge_class"`
	DisplayOrder int       `json:"display_order"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

type Contract struct {
	ID                 int            `json:"id"`
	ServiceRequestID   int            `json:"service_request_id"`
	ContractNumber     string         `json:"contract_number"`
	TotalValue         float64        `json:"total_value"`
	PaymentConditions  string         `json:"payment_conditions"`
	GuaranteeTypeID    int            `json:"guarantee_type_id"`
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
	StatusID           int            `json:"status_id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	
	// Campos relacionados expandidos
	ServiceRequest *ServiceRequest `json:"service_request,omitempty"`
	GuaranteeType  *GuaranteeType  `json:"guarantee_type,omitempty"`
	Status         *ContractStatus `json:"status,omitempty"`
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

// ============================================
// MÉTODOS AUXILIARES
// ============================================

// GetAllGuaranteeTypes retorna todos os tipos de garantia ativos
func (m *ContractModel) GetAllGuaranteeTypes() ([]GuaranteeType, error) {
	query := `SELECT id, code, name, description, requires_custom_text, display_order, active, created_at 
	          FROM guarantee_types WHERE active = true ORDER BY display_order`
	
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []GuaranteeType
	for rows.Next() {
		var gt GuaranteeType
		err := rows.Scan(&gt.ID, &gt.Code, &gt.Name, &gt.Description, &gt.RequiresCustomText, 
			&gt.DisplayOrder, &gt.Active, &gt.CreatedAt)
		if err != nil {
			return nil, err
		}
		types = append(types, gt)
	}
	return types, nil
}

// GetAllContractStatuses retorna todos os status de contrato ativos
func (m *ContractModel) GetAllContractStatuses() ([]ContractStatus, error) {
	query := `SELECT id, code, name, description, color_class, badge_class, display_order, active, created_at 
	          FROM contract_status WHERE active = true ORDER BY display_order`
	
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []ContractStatus
	for rows.Next() {
		var cs ContractStatus
		err := rows.Scan(&cs.ID, &cs.Code, &cs.Name, &cs.Description, &cs.ColorClass, 
			&cs.BadgeClass, &cs.DisplayOrder, &cs.Active, &cs.CreatedAt)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, cs)
	}
	return statuses, nil
}

// GetStatusIDByCode retorna o ID de um status pelo código
func (m *ContractModel) GetStatusIDByCode(code string) (int, error) {
	var id int
	err := m.DB.QueryRow("SELECT id FROM contract_status WHERE code = $1", code).Scan(&id)
	return id, err
}

// GetGuaranteeTypeIDByCode retorna o ID de um tipo de garantia pelo código
func (m *ContractModel) GetGuaranteeTypeIDByCode(code string) (int, error) {
	var id int
	err := m.DB.QueryRow("SELECT id FROM guarantee_types WHERE code = $1", code).Scan(&id)
	return id, err
}

// ============================================
// CRUD DE CONTRATOS
// ============================================

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
	
	// Obter ID do status "RASCUNHO"
	statusID, err := m.GetStatusIDByCode("RASCUNHO")
	if err != nil {
		return fmt.Errorf("erro ao obter status RASCUNHO: %w", err)
	}
	contract.StatusID = statusID

	query := `
		INSERT INTO contracts (
			service_request_id, contract_number, total_value, payment_conditions,
			guarantee_type_id, guarantee_custom, client_requirements, materials_used,
			additional_notes, status_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	return m.DB.QueryRow(
		query,
		contract.ServiceRequestID, contract.ContractNumber, contract.TotalValue,
		contract.PaymentConditions, contract.GuaranteeTypeID, contract.GuaranteeCustom,
		contract.ClientRequirements, contract.MaterialsUsed, contract.AdditionalNotes,
		contract.StatusID,
	).Scan(&contract.ID, &contract.CreatedAt, &contract.UpdatedAt)
}

// Update atualiza um contrato (apenas se não estiver assinado)
func (m *ContractModel) Update(contract *Contract) error {
	query := `
		UPDATE contracts SET
			total_value = $1, payment_conditions = $2, guarantee_type_id = $3,
			guarantee_custom = $4, client_requirements = $5, materials_used = $6,
			additional_notes = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8 AND client_signed = false AND company_signed = false`

	result, err := m.DB.Exec(
		query,
		contract.TotalValue, contract.PaymentConditions, contract.GuaranteeTypeID,
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
	
	// Obter ID do novo status
	newStatusID, err := m.GetStatusIDByCode("AGUARDANDO_ASSINATURAS")
	if err != nil {
		return fmt.Errorf("erro ao obter status AGUARDANDO_ASSINATURAS: %w", err)
	}
	
	// Obter ID do status RASCUNHO
	draftStatusID, err := m.GetStatusIDByCode("RASCUNHO")
	if err != nil {
		return fmt.Errorf("erro ao obter status RASCUNHO: %w", err)
	}
	
	// Verificar se o contrato existe e está em rascunho
	var currentStatusID int
	checkQuery := `SELECT status_id FROM contracts WHERE id = $1`
	
	err = m.DB.QueryRow(checkQuery, contractID).Scan(&currentStatusID)
	
	if err == sql.ErrNoRows {
		log.Printf("❌ [MODEL] Contrato #%d NÃO ENCONTRADO no banco", contractID)
		return fmt.Errorf("contrato #%d não encontrado", contractID)
	}
	if err != nil {
		log.Printf("❌ [MODEL] ERRO ao buscar contrato: %v", err)
		return fmt.Errorf("erro ao buscar contrato: %w", err)
	}
	
	log.Printf("✅ [MODEL] Contrato encontrado. Status ID atual: %d", currentStatusID)
	
	if currentStatusID != draftStatusID {
		log.Printf("⚠️ [MODEL] Status incorreto! Esperado: %d (RASCUNHO), Recebido: %d", draftStatusID, currentStatusID)
		return fmt.Errorf("contrato não pode ser enviado - não está em RASCUNHO")
	}
	
	// Atualizar status
	updateQuery := `UPDATE contracts 
	                SET status_id = $1, updated_at = CURRENT_TIMESTAMP 
	                WHERE id = $2`
	
	log.Printf("📝 [MODEL] Atualizando status de %d para %d (AGUARDANDO_ASSINATURAS)", currentStatusID, newStatusID)
	
	result, err := m.DB.Exec(updateQuery, newStatusID, contractID)
	if err != nil {
		log.Printf("❌ [MODEL] ERRO ao executar UPDATE: %v", err)
		return fmt.Errorf("erro ao atualizar status: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ [MODEL] Linhas afetadas: %d", rowsAffected)
	log.Printf("✅ [MODEL] SendForSignature concluído com sucesso!")
	
	return nil
}

// SignByClient assina o contrato pelo cliente
func (m *ContractModel) SignByClient(contractID int, signature string) error {
	// Obter ID do status "AGUARDANDO_ASSINATURAS"
	waitingStatusID, err := m.GetStatusIDByCode("AGUARDANDO_ASSINATURAS")
	if err != nil {
		return fmt.Errorf("erro ao obter status: %w", err)
	}
	
	// Verificar se o contrato existe e está aguardando assinaturas
	var statusID int
	var clientSigned bool
	checkQuery := `SELECT status_id, client_signed FROM contracts WHERE id = $1`
	err = m.DB.QueryRow(checkQuery, contractID).Scan(&statusID, &clientSigned)
	
	if err == sql.ErrNoRows {
		return fmt.Errorf("contrato não encontrado")
	}
	if err != nil {
		return fmt.Errorf("erro ao verificar contrato: %w", err)
	}
	
	if statusID != waitingStatusID {
		return fmt.Errorf("contrato não está aguardando assinaturas")
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

// SignByCompany assina o contrato pela empresa
func (m *ContractModel) SignByCompany(contractID int, signature string) error {
	// Obter ID do status "AGUARDANDO_ASSINATURAS"
	waitingStatusID, err := m.GetStatusIDByCode("AGUARDANDO_ASSINATURAS")
	if err != nil {
		return fmt.Errorf("erro ao obter status: %w", err)
	}
	
	// Verificar se o contrato existe e está aguardando assinaturas
	var statusID int
	var companySigned bool
	checkQuery := `SELECT status_id, company_signed FROM contracts WHERE id = $1`
	err = m.DB.QueryRow(checkQuery, contractID).Scan(&statusID, &companySigned)
	
	if err == sql.ErrNoRows {
		return fmt.Errorf("contrato não encontrado")
	}
	if err != nil {
		return fmt.Errorf("erro ao verificar contrato: %w", err)
	}
	
	if statusID != waitingStatusID {
		return fmt.Errorf("contrato não está aguardando assinaturas")
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
		// Obter ID do status "ASSINADO"
		var signedStatusID int
		err := tx.QueryRow("SELECT id FROM contract_status WHERE code = 'ASSINADO'").Scan(&signedStatusID)
		if err == nil {
			tx.Exec("UPDATE contracts SET status_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", 
				signedStatusID, contractID)
		}
	}
}

// GetByID busca um contrato pelo ID com dados relacionados
func (m *ContractModel) GetByID(id int) (*Contract, error) {
	contract := &Contract{}
	query := `
		SELECT c.id, c.service_request_id, c.contract_number, c.total_value, c.payment_conditions,
			c.guarantee_type_id, c.guarantee_custom, c.client_requirements, c.materials_used,
			c.additional_notes, c.client_signed, c.client_signed_at, c.client_signature,
			c.company_signed, c.company_signed_at, c.company_signature, c.status_id, 
			c.created_at, c.updated_at,
			gt.id, gt.code, gt.name, gt.description, gt.requires_custom_text,
			cs.id, cs.code, cs.name, cs.description, cs.color_class, cs.badge_class
		FROM contracts c
		LEFT JOIN guarantee_types gt ON c.guarantee_type_id = gt.id
		LEFT JOIN contract_status cs ON c.status_id = cs.id
		WHERE c.id = $1`

	guaranteeType := &GuaranteeType{}
	status := &ContractStatus{}

	err := m.DB.QueryRow(query, id).Scan(
		&contract.ID, &contract.ServiceRequestID, &contract.ContractNumber,
		&contract.TotalValue, &contract.PaymentConditions, &contract.GuaranteeTypeID,
		&contract.GuaranteeCustom, &contract.ClientRequirements, &contract.MaterialsUsed,
		&contract.AdditionalNotes, &contract.ClientSigned, &contract.ClientSignedAt,
		&contract.ClientSignature, &contract.CompanySigned, &contract.CompanySignedAt,
		&contract.CompanySignature, &contract.StatusID, &contract.CreatedAt, &contract.UpdatedAt,
		&guaranteeType.ID, &guaranteeType.Code, &guaranteeType.Name, &guaranteeType.Description,
		&guaranteeType.RequiresCustomText,
		&status.ID, &status.Code, &status.Name, &status.Description, &status.ColorClass, &status.BadgeClass,
	)
	
	if err != nil {
		return nil, err
	}

	contract.GuaranteeType = guaranteeType
	contract.Status = status
	
	return contract, nil
}

// GetByServiceRequestID busca contrato pela solicitação
func (m *ContractModel) GetByServiceRequestID(serviceRequestID int) (*Contract, error) {
	contract := &Contract{}
	query := `
		SELECT c.id, c.service_request_id, c.contract_number, c.total_value, c.payment_conditions,
			c.guarantee_type_id, c.guarantee_custom, c.client_requirements, c.materials_used,
			c.additional_notes, c.client_signed, c.client_signed_at, c.client_signature,
			c.company_signed, c.company_signed_at, c.company_signature, c.status_id, 
			c.created_at, c.updated_at,
			gt.code, cs.code
		FROM contracts c
		LEFT JOIN guarantee_types gt ON c.guarantee_type_id = gt.id
		LEFT JOIN contract_status cs ON c.status_id = cs.id
		WHERE c.service_request_id = $1`

	var guaranteeCode, statusCode string

	err := m.DB.QueryRow(query, serviceRequestID).Scan(
		&contract.ID, &contract.ServiceRequestID, &contract.ContractNumber,
		&contract.TotalValue, &contract.PaymentConditions, &contract.GuaranteeTypeID,
		&contract.GuaranteeCustom, &contract.ClientRequirements, &contract.MaterialsUsed,
		&contract.AdditionalNotes, &contract.ClientSigned, &contract.ClientSignedAt,
		&contract.ClientSignature, &contract.CompanySigned, &contract.CompanySignedAt,
		&contract.CompanySignature, &contract.StatusID, &contract.CreatedAt, &contract.UpdatedAt,
		&guaranteeCode, &statusCode,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Criar objetos simplificados
	contract.GuaranteeType = &GuaranteeType{Code: guaranteeCode}
	contract.Status = &ContractStatus{Code: statusCode}
	
	return contract, nil
}

// GetAllWithDetails busca todos os contratos com detalhes
func (m *ContractModel) GetAllWithDetails(statusCode string, limit, offset int) ([]Contract, int, error) {
	baseQuery := ` FROM contracts c 
	               LEFT JOIN contract_status cs ON c.status_id = cs.id
	               LEFT JOIN guarantee_types gt ON c.guarantee_type_id = gt.id
	               WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if statusCode != "" {
		baseQuery += fmt.Sprintf(" AND cs.code = $%d", argPos)
		args = append(args, statusCode)
		argPos++
	}

	var total int
	countQuery := "SELECT COUNT(*)" + baseQuery
	m.DB.QueryRow(countQuery, args...).Scan(&total)

	selectQuery := fmt.Sprintf(`
		SELECT c.id, c.service_request_id, c.contract_number, c.total_value, c.payment_conditions,
			c.guarantee_type_id, c.guarantee_custom, c.client_requirements, c.materials_used,
			c.additional_notes, c.client_signed, c.client_signed_at, c.client_signature,
			c.company_signed, c.company_signed_at, c.company_signature, c.status_id, 
			c.created_at, c.updated_at,
			cs.code, cs.name, cs.badge_class,
			gt.code, gt.name
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
		var statusCode, statusName, badgeClass, gtCode, gtName string
		
		err := rows.Scan(
			&c.ID, &c.ServiceRequestID, &c.ContractNumber, &c.TotalValue, &c.PaymentConditions,
			&c.GuaranteeTypeID, &c.GuaranteeCustom, &c.ClientRequirements, &c.MaterialsUsed,
			&c.AdditionalNotes, &c.ClientSigned, &c.ClientSignedAt, &c.ClientSignature,
			&c.CompanySigned, &c.CompanySignedAt, &c.CompanySignature, &c.StatusID, 
			&c.CreatedAt, &c.UpdatedAt,
			&statusCode, &statusName, &badgeClass,
			&gtCode, &gtName,
		)
		if err != nil {
			return nil, 0, err
		}
		
		c.Status = &ContractStatus{Code: statusCode, Name: statusName, BadgeClass: badgeClass}
		c.GuaranteeType = &GuaranteeType{Code: gtCode, Name: gtName}
		
		contracts = append(contracts, c)
	}
	return contracts, total, nil
}

// GetAllByUserID busca todos os contratos de um usuário específico
func (m *ContractModel) GetAllByUserID(userID int, statusCode string, limit, offset int) ([]Contract, int, error) {
	baseQuery := ` FROM contracts c 
	               JOIN service_requests sr ON c.service_request_id = sr.id 
	               LEFT JOIN contract_status cs ON c.status_id = cs.id
	               LEFT JOIN guarantee_types gt ON c.guarantee_type_id = gt.id
	               WHERE sr.user_id = $1`
	args := []interface{}{userID}
	argPos := 2

	if statusCode != "" {
		baseQuery += fmt.Sprintf(" AND cs.code = $%d", argPos)
		args = append(args, statusCode)
		argPos++
	}

	var total int
	countQuery := "SELECT COUNT(*)" + baseQuery
	err := m.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	selectQuery := fmt.Sprintf(`
		SELECT c.id, c.service_request_id, c.contract_number, c.total_value, c.payment_conditions,
			c.guarantee_type_id, c.guarantee_custom, c.client_requirements, c.materials_used,
			c.additional_notes, c.client_signed, c.client_signed_at, c.client_signature,
			c.company_signed, c.company_signed_at, c.company_signature, c.status_id, 
			c.created_at, c.updated_at,
			cs.code, cs.name, cs.badge_class,
			gt.code, gt.name
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
		var statusCode, statusName, badgeClass, gtCode, gtName string
		
		err := rows.Scan(
			&c.ID, &c.ServiceRequestID, &c.ContractNumber, &c.TotalValue, &c.PaymentConditions,
			&c.GuaranteeTypeID, &c.GuaranteeCustom, &c.ClientRequirements, &c.MaterialsUsed,
			&c.AdditionalNotes, &c.ClientSigned, &c.ClientSignedAt, &c.ClientSignature,
			&c.CompanySigned, &c.CompanySignedAt, &c.CompanySignature, &c.StatusID, 
			&c.CreatedAt, &c.UpdatedAt,
			&statusCode, &statusName, &badgeClass,
			&gtCode, &gtName,
		)
		if err != nil {
			return nil, 0, err
		}
		
		c.Status = &ContractStatus{Code: statusCode, Name: statusName, BadgeClass: badgeClass}
		c.GuaranteeType = &GuaranteeType{Code: gtCode, Name: gtName}
		
		contracts = append(contracts, c)
	}
	
	return contracts, total, nil
}

// GetPendingForClient busca contratos pendentes de assinatura do cliente
func (m *ContractModel) GetPendingForClient(userID int) ([]Contract, error) {
	// Obter ID do status "AGUARDANDO_ASSINATURAS"
	waitingStatusID, err := m.GetStatusIDByCode("AGUARDANDO_ASSINATURAS")
	if err != nil {
		return nil, err
	}
	
	query := `
		SELECT c.id, c.service_request_id, c.contract_number, c.total_value, c.status_id, c.created_at
		FROM contracts c
		JOIN service_requests sr ON c.service_request_id = sr.id
		WHERE sr.user_id = $1 AND c.status_id = $2 AND c.client_signed = false`

	rows, err := m.DB.Query(query, userID, waitingStatusID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []Contract
	for rows.Next() {
		var c Contract
		rows.Scan(&c.ID, &c.ServiceRequestID, &c.ContractNumber, &c.TotalValue, &c.StatusID, &c.CreatedAt)
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// CanEdit verifica se o contrato pode ser editado
func (m *ContractModel) CanEdit(contractID int) bool {
	var clientSigned, companySigned bool
	m.DB.QueryRow("SELECT client_signed, company_signed FROM contracts WHERE id = $1", contractID).
		Scan(&clientSigned, &companySigned)
	return !clientSigned && !companySigned
}

// AddHistory adiciona registro no histórico
func (m *ContractModel) AddHistory(contractID, userID int, action, fields string) error {
	query := `INSERT INTO contract_history (contract_id, action, changed_by, changed_fields) VALUES ($1, $2, $3, $4)`
	_, err := m.DB.Exec(query, contractID, action, userID, fields)
	return err
}