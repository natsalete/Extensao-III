package models

import (
	"database/sql"
	"time"
)

type ServiceRequest struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	FullName      string    `json:"full_name"`
	ServiceType   string    `json:"service_type"`
	Description   string    `json:"description"`
	CEP           string    `json:"cep"`
	Logradouro    string    `json:"logradouro"`
	Numero        string    `json:"numero"`
	Bairro        string    `json:"bairro"`
	Cidade        string    `json:"cidade"`
	Estado        string    `json:"estado"`
	PreferredDate time.Time `json:"preferred_date"`
	PreferredTime string    `json:"preferred_time"`
	Status        string    `json:"status"` // "SOLICITADA", "CONFIRMADA", "REALIZADA", "CANCELADA"
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	UserName      string    `json:"user_name,omitempty"`
	UserEmail     string    `json:"user_email,omitempty"`
}

type ServiceModel struct {
	DB *sql.DB
}

func NewServiceModel(db *sql.DB) *ServiceModel {
	return &ServiceModel{DB: db}
}

func (m *ServiceModel) Create(service *ServiceRequest) error {
	query := `
		INSERT INTO service_requests (
			user_id, full_name, service_type, description, cep, logradouro, 
			numero, bairro, cidade, estado, preferred_date, preferred_time
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, status, created_at, updated_at`

	return m.DB.QueryRow(
		query, 
		service.UserID, service.FullName, service.ServiceType, service.Description,
		service.CEP, service.Logradouro, service.Numero, service.Bairro,
		service.Cidade, service.Estado, service.PreferredDate, service.PreferredTime,
	).Scan(&service.ID, &service.Status, &service.CreatedAt, &service.UpdatedAt)
}

func (m *ServiceModel) Update(service *ServiceRequest) error {
	query := `
		UPDATE service_requests 
		SET full_name = $1, service_type = $2, description = $3, cep = $4, 
		    logradouro = $5, numero = $6, bairro = $7, cidade = $8, estado = $9, 
		    preferred_date = $10, preferred_time = $11, updated_at = CURRENT_TIMESTAMP
		WHERE id = $12 AND user_id = $13 AND status = 'SOLICITADA'`

	result, err := m.DB.Exec(
		query, 
		service.FullName, service.ServiceType, service.Description, service.CEP,
		service.Logradouro, service.Numero, service.Bairro, service.Cidade,
		service.Estado, service.PreferredDate, service.PreferredTime,
		service.ID, service.UserID,
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
		SET status = 'CANCELADA', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND user_id = $2 AND status = 'SOLICITADA'`

	result, err := m.DB.Exec(query, id, userID)
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

func (m *ServiceModel) GetByUserID(userID int) ([]ServiceRequest, error) {
	query := `
		SELECT id, full_name, service_type, description, cep, logradouro, numero, 
		       bairro, cidade, estado, preferred_date, preferred_time, status, 
		       created_at, updated_at
		FROM service_requests 
		WHERE user_id = $1 
		ORDER BY created_at DESC`

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
			&req.ID, &req.FullName, &req.ServiceType, &req.Description,
			&req.CEP, &req.Logradouro, &req.Numero, &req.Bairro,
			&req.Cidade, &req.Estado, &req.PreferredDate, &preferredTime,
			&req.Status, &req.CreatedAt, &req.UpdatedAt,
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
		SELECT sr.id, sr.user_id, sr.full_name, sr.service_type, sr.description, 
		       sr.cep, sr.logradouro, sr.numero, sr.bairro, sr.cidade, sr.estado,
		       sr.preferred_date, sr.preferred_time, sr.status, sr.created_at, 
		       sr.updated_at, u.name, u.email
		FROM service_requests sr
		JOIN users u ON sr.user_id = u.id
		JOIN user_types ut ON u.user_type_id = ut.id
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
			&req.ID, &req.UserID, &req.FullName, &req.ServiceType, &req.Description,
			&req.CEP, &req.Logradouro, &req.Numero, &req.Bairro, &req.Cidade, &req.Estado,
			&req.PreferredDate, &preferredTime, &req.Status, &req.CreatedAt,
			&req.UpdatedAt, &req.UserName, &req.UserEmail,
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

func (m *ServiceModel) UpdateStatus(id int, status string) error {
	query := "UPDATE service_requests SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	_, err := m.DB.Exec(query, status, id)
	return err
}

func (m *ServiceModel) GetByID(id int) (*ServiceRequest, error) {
	service := &ServiceRequest{}
	var preferredTime sql.NullString
	
	query := `
		SELECT sr.id, sr.user_id, sr.full_name, sr.service_type, sr.description,
		       sr.cep, sr.logradouro, sr.numero, sr.bairro, sr.cidade, sr.estado,
		       sr.preferred_date, sr.preferred_time, sr.status, sr.created_at, 
		       sr.updated_at, u.name, u.email
		FROM service_requests sr
		JOIN users u ON sr.user_id = u.id
		JOIN user_types ut ON u.user_type_id = ut.id
		WHERE sr.id = $1`
	
	err := m.DB.QueryRow(query, id).Scan(
		&service.ID, &service.UserID, &service.FullName, &service.ServiceType,
		&service.Description, &service.CEP, &service.Logradouro, &service.Numero,
		&service.Bairro, &service.Cidade, &service.Estado, &service.PreferredDate,
		&preferredTime, &service.Status, &service.CreatedAt, &service.UpdatedAt,
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
		SELECT id, user_id, full_name, service_type, description, cep, logradouro, 
		       numero, bairro, cidade, estado, preferred_date, preferred_time, 
		       status, created_at, updated_at
		FROM service_requests
		WHERE id = $1 AND user_id = $2`
	
	err := m.DB.QueryRow(query, id, userID).Scan(
		&service.ID, &service.UserID, &service.FullName, &service.ServiceType,
		&service.Description, &service.CEP, &service.Logradouro, &service.Numero,
		&service.Bairro, &service.Cidade, &service.Estado, &service.PreferredDate,
		&preferredTime, &service.Status, &service.CreatedAt, &service.UpdatedAt)
	
	if err != nil {
		return nil, err
	}

	if preferredTime.Valid {
		service.PreferredTime = preferredTime.String
	}
	
	return service, nil
}