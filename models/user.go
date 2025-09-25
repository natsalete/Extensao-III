package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserType struct {
	ID          int       `json:"id"`
	TypeName    string    `json:"type_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type User struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	UserTypeID int       `json:"user_type_id"`
	UserType   string    `json:"user_type"` // Para compatibilidade com código existente
	Phone      string    `json:"phone"`
	Address    string    `json:"address"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserModel struct {
	DB *sql.DB
}

func NewUserModel(db *sql.DB) *UserModel {
	return &UserModel{DB: db}
}

func (m *UserModel) Create(user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Buscar o ID do tipo cliente
	var clienteTypeID int
	err = m.DB.QueryRow("SELECT id FROM user_types WHERE type_name = 'cliente'").Scan(&clienteTypeID)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users (name, email, password, user_type_id, phone, address) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return m.DB.QueryRow(query, user.Name, user.Email, string(hashedPassword), clienteTypeID, user.Phone, user.Address).
		Scan(&user.ID, &user.CreatedAt)
}

func (m *UserModel) GetByEmail(email string) (*User, error) {
	user := &User{}
	query := `
		SELECT u.id, u.name, u.email, u.password, u.user_type_id, ut.type_name, u.phone, u.address, u.created_at 
		FROM users u
		INNER JOIN user_types ut ON u.user_type_id = ut.id
		WHERE u.email = $1`
	
	err := m.DB.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.UserTypeID, 
		&user.UserType, &user.Phone, &user.Address, &user.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (m *UserModel) GetByID(id int) (*User, error) {
	user := &User{}
	query := `
		SELECT u.id, u.name, u.email, u.user_type_id, ut.type_name, u.phone, u.address, u.created_at 
		FROM users u
		INNER JOIN user_types ut ON u.user_type_id = ut.id
		WHERE u.id = $1`
	
	err := m.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.UserTypeID, 
		&user.UserType, &user.Phone, &user.Address, &user.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (m *UserModel) ValidatePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Método para buscar todos os tipos de usuário
func (m *UserModel) GetUserTypes() ([]UserType, error) {
	query := "SELECT id, type_name, description, created_at FROM user_types ORDER BY type_name"
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userTypes []UserType
	for rows.Next() {
		var userType UserType
		err := rows.Scan(&userType.ID, &userType.TypeName, &userType.Description, &userType.CreatedAt)
		if err != nil {
			return nil, err
		}
		userTypes = append(userTypes, userType)
	}

	return userTypes, nil
}

// Método para criar usuário com tipo específico (útil para admin criar outros admins)
func (m *UserModel) CreateWithType(user *User, userTypeName string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Buscar o ID do tipo de usuário
	var userTypeID int
	err = m.DB.QueryRow("SELECT id FROM user_types WHERE type_name = $1", userTypeName).Scan(&userTypeID)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users (name, email, password, user_type_id, phone, address) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return m.DB.QueryRow(query, user.Name, user.Email, string(hashedPassword), userTypeID, user.Phone, user.Address).
		Scan(&user.ID, &user.CreatedAt)
}