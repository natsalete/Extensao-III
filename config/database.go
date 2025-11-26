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

	// Tabela de usuários
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

	// Tabela: Tipos de Serviço
	serviceTypesTable := `
	CREATE TABLE IF NOT EXISTS service_types (
		id SERIAL PRIMARY KEY,
		code VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		icon VARCHAR(50),
		active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Tabela: Status de Solicitação
	requestStatusTable := `
	CREATE TABLE IF NOT EXISTS request_status (
		id SERIAL PRIMARY KEY,
		code VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		color_class VARCHAR(50),
		display_order INTEGER,
		active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// ============================================
	// NOVA TABELA: Tipos de Garantia
	// ============================================
	guaranteeTypesTable := `
	CREATE TABLE IF NOT EXISTS guarantee_types (
		id SERIAL PRIMARY KEY,
		code VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		requires_custom_text BOOLEAN DEFAULT false,
		display_order INTEGER,
		active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// ============================================
	// NOVA TABELA: Status de Contrato
	// ============================================
	contractStatusTable := `
	CREATE TABLE IF NOT EXISTS contract_status (
		id SERIAL PRIMARY KEY,
		code VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		color_class VARCHAR(50),
		badge_class VARCHAR(50),
		display_order INTEGER,
		active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Tabela de solicitações de serviços
	serviceRequestTable := `
	CREATE TABLE IF NOT EXISTS service_requests (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id),
		full_name VARCHAR(200) NOT NULL,
		service_type_id INTEGER REFERENCES service_types(id),
		description TEXT,
		cep VARCHAR(10) NOT NULL,
		logradouro VARCHAR(200) NOT NULL,
		numero VARCHAR(20) NOT NULL,
		bairro VARCHAR(100) NOT NULL,
		cidade VARCHAR(100) NOT NULL,
		estado VARCHAR(2) NOT NULL,
		preferred_date DATE NOT NULL,
		preferred_time TIME NOT NULL,
		status_id INTEGER REFERENCES request_status(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// ============================================
	// TABELA DE CONTRATOS REFATORADA
	// ============================================
	contractsTable := `
	CREATE TABLE IF NOT EXISTS contracts (
		id SERIAL PRIMARY KEY,
		service_request_id INTEGER UNIQUE NOT NULL REFERENCES service_requests(id) ON DELETE CASCADE,
		contract_number VARCHAR(50) UNIQUE NOT NULL,
		total_value DECIMAL(10,2) NOT NULL,
		payment_conditions TEXT NOT NULL,
		guarantee_type_id INTEGER NOT NULL REFERENCES guarantee_types(id),
		guarantee_custom TEXT,
		client_requirements TEXT,
		materials_used TEXT,
		additional_notes TEXT,
		client_signed BOOLEAN DEFAULT false,
		client_signed_at TIMESTAMP,
		client_signature TEXT,
		company_signed BOOLEAN DEFAULT false,
		company_signed_at TIMESTAMP,
		company_signature TEXT,
		status_id INTEGER NOT NULL REFERENCES contract_status(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	contractHistoryTable := `
	CREATE TABLE IF NOT EXISTS contract_history (
		id SERIAL PRIMARY KEY,
		contract_id INTEGER NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
		action VARCHAR(50) NOT NULL,
		changed_by INTEGER REFERENCES users(id),
		changed_fields TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Criar tabelas na ordem correta
	tables := []struct {
		name  string
		query string
	}{
		{"user_types", userTypesTable},
		{"users", userTable},
		{"service_types", serviceTypesTable},
		{"request_status", requestStatusTable},
		{"guarantee_types", guaranteeTypesTable},        // NOVA
		{"contract_status", contractStatusTable},         // NOVA
		{"service_requests", serviceRequestTable},
		{"contracts", contractsTable},
		{"contract_history", contractHistoryTable},
	}

	for _, table := range tables {
		if _, err := DB.Exec(table.query); err != nil {
			log.Fatalf("Error creating %s table: %v", table.name, err)
		}
	}

	// Inserir dados padrão
	insertDefaultUserTypes()
	insertDefaultServiceTypes()
	insertDefaultRequestStatus()
	insertDefaultGuaranteeTypes()  // NOVA
	insertDefaultContractStatus()  // NOVA
	createDefaultAdmin()
}

func insertDefaultUserTypes() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM user_types").Scan(&count)
	
	if count == 0 {
		_, err := DB.Exec(`
			INSERT INTO user_types (type_name, description) VALUES 
			('cliente', 'Cliente padrão do sistema'),
			('gestor', 'Gestor/Administrador do sistema')`)
		if err != nil {
			log.Fatal("Error inserting default user types:", err)
		}
		log.Println("Default user types created")
	}
}

func insertDefaultServiceTypes() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM service_types").Scan(&count)
	
	if count == 0 {
		_, err := DB.Exec(`
			INSERT INTO service_types (code, name, description, icon) VALUES 
			('perfuracao', 'Perfuração de Poços', 'Perfuração de poços artesianos', 'construction'),
			('analise', 'Análise da Água', 'Análise de qualidade da água', 'droplets'),
			('manutencao', 'Manutenção', 'Manutenção de poços existentes', 'wrench')`)
		if err != nil {
			log.Fatal("Error inserting default service types:", err)
		}
		log.Println("Default service types created")
	}
}

func insertDefaultRequestStatus() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM request_status").Scan(&count)
	
	if count == 0 {
		_, err := DB.Exec(`
			INSERT INTO request_status (code, name, description, color_class, display_order) VALUES 
			('SOLICITADA', 'Solicitada', 'Solicitação enviada e aguardando análise', 'status-solicitada', 1),
			('CONFIRMADA', 'Confirmada', 'Vistoria confirmada e agendada', 'status-confirmada', 2),
			('REALIZADA', 'Realizada', 'Vistoria realizada com sucesso', 'status-realizada', 3),
			('CANCELADA', 'Cancelada', 'Solicitação cancelada', 'status-cancelada', 4)`)
		if err != nil {
			log.Fatal("Error inserting default request status:", err)
		}
		log.Println("Default request status created")
	}
}

// ============================================
// NOVA FUNÇÃO: Inserir Tipos de Garantia
// ============================================
func insertDefaultGuaranteeTypes() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM guarantee_types").Scan(&count)
	
	if count == 0 {
		_, err := DB.Exec(`
			INSERT INTO guarantee_types (code, name, description, requires_custom_text, display_order) VALUES 
			('SEGUNDA_TENTATIVA', 'Segunda Tentativa', 'Segunda tentativa sem custo adicional', false, 1),
			('SEM_GARANTIA', 'Sem Garantia', 'Sem garantia adicional', false, 2),
			('PERSONALIZADA', 'Garantia Personalizada', 'Garantia com termos personalizados', true, 3)`)
		if err != nil {
			log.Fatal("Error inserting default guarantee types:", err)
		}
		log.Println("Default guarantee types created")
	}
}

// ============================================
// NOVA FUNÇÃO: Inserir Status de Contrato
// ============================================
func insertDefaultContractStatus() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM contract_status").Scan(&count)
	
	if count == 0 {
		_, err := DB.Exec(`
			INSERT INTO contract_status (code, name, description, color_class, badge_class, display_order) VALUES 
			('RASCUNHO', 'Rascunho', 'Contrato em elaboração', 'text-secondary', 'bg-secondary', 1),
			('AGUARDANDO_ASSINATURAS', 'Aguardando Assinaturas', 'Enviado para assinatura das partes', 'text-warning', 'bg-warning text-dark', 2),
			('ASSINADO', 'Assinado', 'Contrato assinado por ambas as partes', 'text-success', 'bg-success', 3),
			('CANCELADO', 'Cancelado', 'Contrato cancelado', 'text-danger', 'bg-danger', 4)`)
		if err != nil {
			log.Fatal("Error inserting default contract status:", err)
		}
		log.Println("Default contract status created")
	}
}

func createDefaultAdmin() {
	var count int
	DB.QueryRow(`
		SELECT COUNT(*) FROM users u
		INNER JOIN user_types ut ON u.user_type_id = ut.id
		WHERE ut.type_name = 'gestor'`).Scan(&count)
	
	if count == 0 {
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