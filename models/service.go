package models

import (
	"database/sql"
	"time"
)

type ServiceRequest struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	ServiceType string    `json:"service_type"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "pendente", "em_andamento", "concluido"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserName    string    `json:"user_name,omitempty"`
	UserEmail   string    `json:"user_email,omitempty"`
}

type ServiceModel struct {
	DB *sql.DB
}

func NewServiceModel(db *sql.DB) *ServiceModel {
	return &ServiceModel{DB: db}
}

func (m *ServiceModel) Create(service *ServiceRequest) error {
	query := `
		INSERT INTO service_requests (user_id, service_type, description) 
		VALUES ($1, $2, $3)
		RETURNING id, status, created_at, updated_at`

	return m.DB.QueryRow(query, service.UserID, service.ServiceType, service.Description).
		Scan(&service.ID, &service.Status, &service.CreatedAt, &service.UpdatedAt)
}

func (m *ServiceModel) GetByUserID(userID int) ([]ServiceRequest, error) {
	query := `
		SELECT id, service_type, description, status, created_at, updated_at
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
		err := rows.Scan(&req.ID, &req.ServiceType, &req.Description, &req.Status, &req.CreatedAt, &req.UpdatedAt)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (m *ServiceModel) GetAll() ([]ServiceRequest, error) {
	query := `
		SELECT sr.id, sr.user_id, sr.service_type, sr.description, sr.status, 
		       sr.created_at, sr.updated_at, u.name, u.email
		FROM service_requests sr
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
		err := rows.Scan(&req.ID, &req.UserID, &req.ServiceType, &req.Description, 
			&req.Status, &req.CreatedAt, &req.UpdatedAt, &req.UserName, &req.UserEmail)
		if err != nil {
			return nil, err
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
	query := `
		SELECT sr.id, sr.user_id, sr.service_type, sr.description, sr.status, 
		       sr.created_at, sr.updated_at, u.name, u.email
		FROM service_requests sr
		JOIN users u ON sr.user_id = u.id
		WHERE sr.id = $1`
	
	err := m.DB.QueryRow(query, id).Scan(
		&service.ID, &service.UserID, &service.ServiceType, &service.Description,
		&service.Status, &service.CreatedAt, &service.UpdatedAt, &service.UserName, &service.UserEmail)
	
	if err != nil {
		return nil, err
	}
	
	return service, nil
}