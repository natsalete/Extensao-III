package models

import (
	"database/sql"
	"time"
	
	"martins-pocos/constants"
)

type ServiceType struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

type RequestStatus struct {
	ID           int       `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ColorClass   string    `json:"color_class"`
	DisplayOrder int       `json:"display_order"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

type ServiceRequest struct {
	ID              int            `json:"id"`
	UserID          int            `json:"user_id"`
	FullName        string         `json:"full_name"`
	ServiceTypeID   int            `json:"service_type_id"`
	ServiceTypeCode string         `json:"service_type_code,omitempty"`
	ServiceTypeName string         `json:"service_type_name,omitempty"`
	ServiceTypeIcon string         `json:"service_type_icon,omitempty"`
	Description     string         `json:"description"`
	CEP             string         `json:"cep"`
	Logradouro      string         `json:"logradouro"`
	Numero          string         `json:"numero"`
	Bairro          string         `json:"bairro"`
	Cidade          string         `json:"cidade"`
	Estado          string         `json:"estado"`
	PreferredDate   time.Time      `json:"preferred_date"`
	PreferredTime   string         `json:"preferred_time"`
	StatusID        int            `json:"status_id"`
	StatusCode      string         `json:"status_code,omitempty"`
	StatusName      string         `json:"status_name,omitempty"`
	StatusColor     string         `json:"status_color,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	UserName        string         `json:"user_name,omitempty"`
	UserEmail       string         `json:"user_email,omitempty"`
}

type ServiceModel struct {
	DB *sql.DB
}

func NewServiceModel(db *sql.DB) *ServiceModel {
	return &ServiceModel{DB: db}
}

// ==================== Service Types Methods ====================

func (m *ServiceModel) GetAllServiceTypes() ([]ServiceType, error) {
	query := `SELECT id, code, name, description, icon, active, created_at 
	          FROM service_types WHERE active = true ORDER BY name`
	
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []ServiceType
	for rows.Next() {
		var t ServiceType
		err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.Description, &t.Icon, &t.Active, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, nil
}

func (m *ServiceModel) GetServiceTypeByCode(code string) (*ServiceType, error) {
	var t ServiceType
	query := `SELECT id, code, name, description, icon, active, created_at 
	          FROM service_types WHERE code = $1 AND active = true`
	
	err := m.DB.QueryRow(query, code).Scan(&t.ID, &t.Code, &t.Name, &t.Description, &t.Icon, &t.Active, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ==================== Request Status Methods ====================

func (m *ServiceModel) GetAllRequestStatus() ([]RequestStatus, error) {
	query := `SELECT id, code, name, description, color_class, display_order, active, created_at 
	          FROM request_status WHERE active = true ORDER BY display_order`
	
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []RequestStatus
	for rows.Next() {
		var s RequestStatus
		err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.Description, &s.ColorClass, &s.DisplayOrder, &s.Active, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}

func (m *ServiceModel) GetRequestStatusByID(id int) (*RequestStatus, error) {
	var s RequestStatus
	query := `SELECT id, code, name, description, color_class, display_order, active, created_at 
	          FROM request_status WHERE id = $1 AND active = true`
	
	err := m.DB.QueryRow(query, id).Scan(&s.ID, &s.Code, &s.Name, &s.Description, &s.ColorClass, &s.DisplayOrder, &s.Active, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ==================== Service Request CRUD Methods ====================

func (m *ServiceModel) Create(service *ServiceRequest) error {
	query := `
		INSERT INTO service_requests (
			user_id, full_name, service_type_id, description, cep, logradouro, 
			numero, bairro, cidade, estado, preferred_date, preferred_time, status_id
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, status_id, created_at, updated_at`

	return m.DB.QueryRow(
		query, 
		service.UserID, service.FullName, service.ServiceTypeID, service.Description,
		service.CEP, service.Logradouro, service.Numero, service.Bairro,
		service.Cidade, service.Estado, service.PreferredDate, service.PreferredTime, 
		constants.StatusSolicitada,
	).Scan(&service.ID, &service.StatusID, &service.CreatedAt, &service.UpdatedAt)
}

func (m *ServiceModel) Update(service *ServiceRequest) error {
	query := `
		UPDATE service_requests 
		SET full_name = $1, service_type_id = $2, description = $3, cep = $4, 
		    logradouro = $5, numero = $6, bairro = $7, cidade = $8, estado = $9, 
		    preferred_date = $10, preferred_time = $11, updated_at = CURRENT_TIMESTAMP
		WHERE id = $12 AND user_id = $13 AND status_id = $14`

	result, err := m.DB.Exec(
		query, 
		service.FullName, service.ServiceTypeID, service.Description, service.CEP,
		service.Logradouro, service.Numero, service.Bairro, service.Cidade,
		service.Estado, service.PreferredDate, service.PreferredTime,
		service.ID, service.UserID, constants.StatusSolicitada,
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

func (m *ServiceModel) Cancel(id, userID int) error {
	query := `
		UPDATE service_requests 
		SET status_id = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND user_id = $3 AND status_id = $4`

	result, err := m.DB.Exec(query, constants.StatusCancelada, id, userID, constants.StatusSolicitada)
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

func (m *ServiceModel) UpdateStatusByID(requestID, statusID int) error {
	query := `UPDATE service_requests 
	          SET status_id = $1, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $2`
	_, err := m.DB.Exec(query, statusID, requestID)
	return err
}

// ==================== Query Methods ====================

func (m *ServiceModel) GetByUserID(userID int) ([]ServiceRequest, error) {
	query := `
		SELECT sr.id, sr.full_name, sr.service_type_id, st.code, st.name, st.icon,
		       sr.description, sr.cep, sr.logradouro, sr.numero, 
		       sr.bairro, sr.cidade, sr.estado, sr.preferred_date, sr.preferred_time,
		       sr.status_id, rs.code, rs.name, rs.color_class,
		       sr.created_at, sr.updated_at
		FROM service_requests sr
		JOIN service_types st ON sr.service_type_id = st.id
		JOIN request_status rs ON sr.status_id = rs.id
		WHERE sr.user_id = $1 
		ORDER BY sr.created_at DESC`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []ServiceRequest
	for rows.Next() {
		var req ServiceRequest
		var preferredTime sql.NullString
		
		err := rows.Scan(
			&req.ID, &req.FullName, &req.ServiceTypeID, &req.ServiceTypeCode, 
			&req.ServiceTypeName, &req.ServiceTypeIcon, &req.Description,
			&req.CEP, &req.Logradouro, &req.Numero, &req.Bairro,
			&req.Cidade, &req.Estado, &req.PreferredDate, &preferredTime,
			&req.StatusID, &req.StatusCode, &req.StatusName, &req.StatusColor,
			&req.CreatedAt, &req.UpdatedAt,
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

func (m *ServiceModel) GetAll() ([]ServiceRequest, error) {
	query := `
		SELECT sr.id, sr.user_id, sr.full_name, sr.service_type_id, st.code, st.name, st.icon,
		       sr.description, sr.cep, sr.logradouro, sr.numero, sr.bairro, sr.cidade, sr.estado,
		       sr.preferred_date, sr.preferred_time, sr.status_id, rs.code, rs.name, rs.color_class,
		       sr.created_at, sr.updated_at, u.name, u.email
		FROM service_requests sr
		JOIN service_types st ON sr.service_type_id = st.id
		JOIN request_status rs ON sr.status_id = rs.id
		JOIN users u ON sr.user_id = u.id
		ORDER BY sr.created_at DESC`

	rows, err := m.DB.Query(query)
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

func (m *ServiceModel) GetByID(id int) (*ServiceRequest, error) {
	service := &ServiceRequest{}
	var preferredTime sql.NullString
	
	query := `
		SELECT sr.id, sr.user_id, sr.full_name, sr.service_type_id, st.code, st.name, st.icon,
		       sr.description, sr.cep, sr.logradouro, sr.numero, sr.bairro, sr.cidade, sr.estado,
		       sr.preferred_date, sr.preferred_time, sr.status_id, rs.code, rs.name, rs.color_class,
		       sr.created_at, sr.updated_at, u.name, u.email
		FROM service_requests sr
		JOIN service_types st ON sr.service_type_id = st.id
		JOIN request_status rs ON sr.status_id = rs.id
		JOIN users u ON sr.user_id = u.id
		WHERE sr.id = $1`
	
	err := m.DB.QueryRow(query, id).Scan(
		&service.ID, &service.UserID, &service.FullName, &service.ServiceTypeID,
		&service.ServiceTypeCode, &service.ServiceTypeName, &service.ServiceTypeIcon,
		&service.Description, &service.CEP, &service.Logradouro, &service.Numero,
		&service.Bairro, &service.Cidade, &service.Estado, &service.PreferredDate,
		&preferredTime, &service.StatusID, &service.StatusCode, &service.StatusName,
		&service.StatusColor, &service.CreatedAt, &service.UpdatedAt,
		&service.UserName, &service.UserEmail)
	
	if err != nil {
		return nil, err
	}

	if preferredTime.Valid {
		service.PreferredTime = preferredTime.String
	}
	
	return service, nil
}

func (m *ServiceModel) GetByIDAndUser(id, userID int) (*ServiceRequest, error) {
	service := &ServiceRequest{}
	var preferredTime sql.NullString
	
	query := `
		SELECT sr.id, sr.user_id, sr.full_name, sr.service_type_id, st.code, st.name, st.icon,
		       sr.description, sr.cep, sr.logradouro, sr.numero, sr.bairro, sr.cidade, sr.estado,
		       sr.preferred_date, sr.preferred_time, sr.status_id, rs.code, rs.name, rs.color_class,
		       sr.created_at, sr.updated_at
		FROM service_requests sr
		JOIN service_types st ON sr.service_type_id = st.id
		JOIN request_status rs ON sr.status_id = rs.id
		WHERE sr.id = $1 AND sr.user_id = $2`
	
	err := m.DB.QueryRow(query, id, userID).Scan(
		&service.ID, &service.UserID, &service.FullName, &service.ServiceTypeID,
		&service.ServiceTypeCode, &service.ServiceTypeName, &service.ServiceTypeIcon,
		&service.Description, &service.CEP, &service.Logradouro, &service.Numero,
		&service.Bairro, &service.Cidade, &service.Estado, &service.PreferredDate,
		&preferredTime, &service.StatusID, &service.StatusCode, &service.StatusName,
		&service.StatusColor, &service.CreatedAt, &service.UpdatedAt)
	
	if err != nil {
		return nil, err
	}

	if preferredTime.Valid {
		service.PreferredTime = preferredTime.String
	}
	
	return service, nil
}