package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	UserType  string    `json:"user_type"` // "cliente" ou "gestor"
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
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

	query := `
		INSERT INTO users (name, email, password, user_type, phone, address) 
		VALUES ($1, $2, $3, 'cliente', $4, $5)
		RETURNING id, created_at`

	return m.DB.QueryRow(query, user.Name, user.Email, string(hashedPassword), user.Phone, user.Address).
		Scan(&user.ID, &user.CreatedAt)
}

func (m *UserModel) GetByEmail(email string) (*User, error) {
	user := &User{}
	query := "SELECT id, name, email, password, user_type, phone, address, created_at FROM users WHERE email = $1"
	
	err := m.DB.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.UserType, 
		&user.Phone, &user.Address, &user.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (m *UserModel) GetByID(id int) (*User, error) {
	user := &User{}
	query := "SELECT id, name, email, user_type, phone, address, created_at FROM users WHERE id = $1"
	
	err := m.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.UserType, 
		&user.Phone, &user.Address, &user.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (m *UserModel) ValidatePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}