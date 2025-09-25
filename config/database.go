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
	// Tabela de tipos de usuário
	userTypesTable := `
	CREATE TABLE IF NOT EXISTS user_types (
		id SERIAL PRIMARY KEY,
		type_name VARCHAR(20) UNIQUE NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Tabela de usuários refatorada
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		user_type_id INTEGER NOT NULL REFERENCES user_types(id),
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

	// Criar tabelas na ordem correta
	if _, err := DB.Exec(userTypesTable); err != nil {
		log.Fatal("Error creating user_types table:", err)
	}

	if _, err := DB.Exec(userTable); err != nil {
		log.Fatal("Error creating users table:", err)
	}

	if _, err := DB.Exec(serviceRequestTable); err != nil {
		log.Fatal("Error creating service_requests table:", err)
	}

	// Inserir tipos de usuário padrão
	insertDefaultUserTypes()
	// Criar administrador padrão
	createDefaultAdmin()
}

func insertDefaultUserTypes() {
	// Verificar se os tipos já existem
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM user_types").Scan(&count)
	
	if count == 0 {
		// Inserir tipos de usuário
		_, err := DB.Exec(`
			INSERT INTO user_types (type_name, description) VALUES 
			('cliente', 'Cliente padrão do sistema'),
			('gestor', 'Gestor/Administrador do sistema')`)
		if err != nil {
			log.Fatal("Error inserting default user types:", err)
		} else {
			log.Println("Default user types created: cliente, gestor")
		}
	}
}

func createDefaultAdmin() {
	// Verificar se já existe um gestor
	var count int
	DB.QueryRow(`
		SELECT COUNT(*) FROM users u
		INNER JOIN user_types ut ON u.user_type_id = ut.id
		WHERE ut.type_name = 'gestor'`).Scan(&count)
	
	if count == 0 {
		// Buscar o ID do tipo gestor
		var gestorTypeID int
		err := DB.QueryRow("SELECT id FROM user_types WHERE type_name = 'gestor'").Scan(&gestorTypeID)
		if err != nil {
			log.Fatal("Error finding gestor type ID:", err)
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		_, err = DB.Exec(`
			INSERT INTO users (name, email, password, user_type_id, phone) 
			VALUES ($1, $2, $3, $4, $5)`,
			"Administrador", "admin@martinspocos.com", string(hashedPassword), gestorTypeID, "(34) 9999-9999")
		if err != nil {
			log.Println("Error creating default admin:", err)
		} else {
			log.Println("Default admin created: admin@martinspocos.com / admin123")
		}
	}
}