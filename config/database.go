package config

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func InitDB() {
	var err error
	connStr := "user=postgres dbname=martins_pocos sslmode=disable password=123456"
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}

	createTables()
}

func GetDB() *sql.DB {
	return DB
}

func createTables() {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('cliente', 'gestor')),
		phone VARCHAR(20),
		address TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	serviceRequestTable := `
	CREATE TABLE IF NOT EXISTS service_requests (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id),
		service_type VARCHAR(50) NOT NULL,
		description TEXT,
		status VARCHAR(20) DEFAULT 'pendente' CHECK (status IN ('pendente', 'em_andamento', 'concluido')),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := DB.Exec(userTable); err != nil {
		log.Fatal("Error creating users table:", err)
	}

	if _, err := DB.Exec(serviceRequestTable); err != nil {
		log.Fatal("Error creating service_requests table:", err)
	}

	createDefaultAdmin()
}

func createDefaultAdmin() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM users WHERE user_type = 'gestor'").Scan(&count)
	
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		_, err := DB.Exec(`
			INSERT INTO users (name, email, password, user_type, phone) 
			VALUES ($1, $2, $3, $4, $5)`,
			"Administrador", "admin@martinspocos.com", string(hashedPassword), "gestor", "(34) 9999-9999")
		if err != nil {
			log.Println("Error creating default admin:", err)
		} else {
			log.Println("Default admin created: admin@martinspocos.com / admin123")
		}
	}
}