package models

import (
	"database/sql"
	"strconv"
)

// GetAllWithFilters busca todas as solicitações com filtros (para admin)
func (m *ServiceModel) GetAllWithFilters(statusFilter, serviceTypeFilter, searchQuery string, limit, offset int) ([]ServiceRequest, int, error) {
	baseQuery := `
		FROM service_requests sr
		JOIN service_types st ON sr.service_type_id = st.id
		JOIN request_status rs ON sr.status_id = rs.id
		JOIN users u ON sr.user_id = u.id
		WHERE 1=1`
	
	args := []interface{}{}
	argPos := 1
	
	// Filtro de status
	if statusFilter != "" {
		baseQuery += " AND rs.code = $" + strconv.Itoa(argPos)
		args = append(args, statusFilter)
		argPos++
	}
	
	// Filtro de tipo de serviço
	if serviceTypeFilter != "" {
		baseQuery += " AND st.code = $" + strconv.Itoa(argPos)
		args = append(args, serviceTypeFilter)
		argPos++
	}
	
	// Filtro de busca (nome, cidade, email)
	if searchQuery != "" {
		baseQuery += " AND (LOWER(sr.full_name) LIKE LOWER($" + strconv.Itoa(argPos) + 
					  ") OR LOWER(sr.cidade) LIKE LOWER($" + strconv.Itoa(argPos) + 
					  ") OR LOWER(u.email) LIKE LOWER($" + strconv.Itoa(argPos) + "))"
		args = append(args, "%"+searchQuery+"%")
		argPos++
	}
	
	// Contar total
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int
	err := m.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Query completa
	selectQuery := `
		SELECT sr.id, sr.user_id, sr.full_name, sr.service_type_id, st.code, st.name, st.icon,
		       sr.description, sr.cep, sr.logradouro, sr.numero, sr.bairro, sr.cidade, sr.estado,
		       sr.preferred_date, sr.preferred_time, sr.status_id, rs.code, rs.name, rs.color_class,
		       sr.created_at, sr.updated_at, u.name, u.email
		` + baseQuery + `
		ORDER BY sr.created_at DESC
		LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)
	
	args = append(args, limit, offset)
	
	rows, err := m.DB.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var requests []ServiceRequest
	for rows.Next() {
		var req ServiceRequest
		var preferredTime sql.NullString
		
		err := rows.Scan(
			&req.ID, &req.UserID, &req.FullName, &req.ServiceTypeID, &req.ServiceTypeCode,
			&req.ServiceTypeName, &req.ServiceTypeIcon, &req.Description,
			&req.CEP, &req.Logradouro, &req.Numero, &req.Bairro, &req.Cidade, &req.Estado,
			&req.PreferredDate, &preferredTime, &req.StatusID, &req.StatusCode,
			&req.StatusName, &req.StatusColor, &req.CreatedAt, &req.UpdatedAt,
			&req.UserName, &req.UserEmail,
		)
		if err != nil {
			return nil, 0, err
		}
		
		if preferredTime.Valid {
			req.PreferredTime = preferredTime.String
		}
		
		requests = append(requests, req)
	}
	
	return requests, totalCount, nil
}

// AdminUpdate atualiza uma solicitação (admin pode atualizar status também)
func (m *ServiceModel) AdminUpdate(service *ServiceRequest) error {
	query := `
		UPDATE service_requests 
		SET full_name = $1, service_type_id = $2, description = $3, cep = $4, 
		    logradouro = $5, numero = $6, bairro = $7, cidade = $8, estado = $9, 
		    preferred_date = $10, preferred_time = $11, status_id = $12, updated_at = CURRENT_TIMESTAMP
		WHERE id = $13`

	result, err := m.DB.Exec(
		query, 
		service.FullName, service.ServiceTypeID, service.Description, service.CEP,
		service.Logradouro, service.Numero, service.Bairro, service.Cidade,
		service.Estado, service.PreferredDate, service.PreferredTime, service.StatusID,
		service.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete deleta uma solicitação (apenas admin)
func (m *ServiceModel) Delete(requestID int) error {
	query := `DELETE FROM service_requests WHERE id = $1`
	
	result, err := m.DB.Exec(query, requestID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetStatusStats retorna estatísticas por status
func (m *ServiceModel) GetStatusStats() (map[string]int, error) {
	query := `
		SELECT rs.code, COUNT(sr.id) as count
		FROM request_status rs
		LEFT JOIN service_requests sr ON sr.status_id = rs.id
		GROUP BY rs.code, rs.display_order
		ORDER BY rs.display_order`
	
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var code string
		var count int
		if err := rows.Scan(&code, &count); err != nil {
			return nil, err
		}
		stats[code] = count
	}

	return stats, nil
}

// GetRecentRequests retorna as N últimas solicitações
func (m *ServiceModel) GetRecentRequests(limit int) ([]ServiceRequest, error) {
	query := `
		SELECT sr.id, sr.user_id, sr.full_name, sr.service_type_id, st.code, st.name, st.icon,
		       sr.description, sr.cep, sr.logradouro, sr.numero, sr.bairro, sr.cidade, sr.estado,
		       sr.preferred_date, sr.preferred_time, sr.status_id, rs.code, rs.name, rs.color_class,
		       sr.created_at, sr.updated_at, u.name, u.email
		FROM service_requests sr
		JOIN service_types st ON sr.service_type_id = st.id
		JOIN request_status rs ON sr.status_id = rs.id
		JOIN users u ON sr.user_id = u.id
		ORDER BY sr.created_at DESC
		LIMIT $1`

	rows, err := m.DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []ServiceRequest
	for rows.Next() {
		var req ServiceRequest
		var preferredTime sql.NullString
		
		err := rows.Scan(
			&req.ID, &req.UserID, &req.FullName, &req.ServiceTypeID, &req.ServiceTypeCode,
			&req.ServiceTypeName, &req.ServiceTypeIcon, &req.Description,
			&req.CEP, &req.Logradouro, &req.Numero, &req.Bairro, &req.Cidade, &req.Estado,
			&req.PreferredDate, &preferredTime, &req.StatusID, &req.StatusCode,
			&req.StatusName, &req.StatusColor, &req.CreatedAt, &req.UpdatedAt,
			&req.UserName, &req.UserEmail,
		)
		if err != nil {
			return nil, err
		}

		if preferredTime.Valid {
			req.PreferredTime = preferredTime.String
		}

		requests = append(requests, req)
	}

	return requests, nil
}