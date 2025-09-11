-- Criar o banco de dados
CREATE DATABASE martins_pocos;

-- Conectar ao banco
\c martins_pocos;

-- Criar tabela de usuários
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('cliente', 'gestor')),
    phone VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Criar tabela de solicitações de serviço
CREATE TABLE IF NOT EXISTS service_requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    service_type VARCHAR(50) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'pendente' CHECK (status IN ('pendente', 'em_andamento', 'concluido')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Criar usuário administrador padrão
-- Senha: admin123 (hash gerado pelo bcrypt)
INSERT INTO users (name, email, password, user_type, phone) 
VALUES (
    'Administrador', 
    'admin@martinspocos.com', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 
    'gestor', 
    '(34) 9999-9999'
);

-- Criar alguns dados de exemplo
INSERT INTO users (name, email, password, user_type, phone, address) 
VALUES (
    'João Silva', 
    'joao@email.com', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 
    'cliente', 
    '(34) 9888-7777',
    'Rua das Flores, 123, Centro, Monte Carmelo - MG'
);

INSERT INTO users (name, email, password, user_type, phone, address) 
VALUES (
    'Maria Santos', 
    'maria@email.com', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 
    'cliente', 
    '(34) 9777-6666',
    'Fazenda Boa Vista, Zona Rural, Monte Carmelo - MG'
);

-- Criar algumas solicitações de exemplo
INSERT INTO service_requests (user_id, service_type, description, status) 
VALUES (
    2, 
    'perfuracao', 
    'Preciso de um poço artesiano na minha propriedade rural. O terreno é plano e tem fácil acesso. Urgência alta devido à seca.', 
    'em_andamento'
);

INSERT INTO service_requests (user_id, service_type, description, status) 
VALUES (
    3, 
    'analise', 
    'Tenho um poço há 5 anos e gostaria de fazer uma análise da qualidade da água para consumo humano.', 
    'pendente'
);

INSERT INTO service_requests (user_id, service_type, description, status) 
VALUES (
    2, 
    'manutencao', 
    'Meu poço está com baixa vazão. Preciso de manutenção preventiva e limpeza.', 
    'concluido'
);

-- Criar índices para melhor performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_type ON users(user_type);
CREATE INDEX idx_service_requests_user_id ON service_requests(user_id);
CREATE INDEX idx_service_requests_status ON service_requests(status);
CREATE INDEX idx_service_requests_created_at ON service_requests(created_at);