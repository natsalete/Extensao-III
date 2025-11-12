package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type WhatsAppService struct {
	APIKey      string
	InstanceID  string
	BaseURL     string
}

// Estrutura de resposta da Z-API para Status
type ZAPIStatusResponse struct {
	Connected bool   `json:"connected"`
	Session   string `json:"session"`
	Error     string `json:"error"`
	Smartphon struct {
		Connected bool   `json:"connected"`
		Number    string `json:"number"`
	} `json:"smartphon"`
}

// Estrutura de resposta da Z-API para envio de mensagem
type ZAPIResponse struct {
	ZaapId  string `json:"zaapId"`
	MessageId string `json:"messageId"`
	Id      string `json:"id"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func NewWhatsAppService() *WhatsAppService {
	return &WhatsAppService{
		APIKey:     os.Getenv("WHATSAPP_API_KEY"),
		InstanceID: os.Getenv("WHATSAPP_INSTANCE_ID"),
		BaseURL:    "https://api.z-api.io",
	}
}

// CheckConnection verifica se a instância está conectada
func (s *WhatsAppService) CheckConnection() (bool, error) {
	endpoint := fmt.Sprintf("%s/instances/%s/token/%s/status",
		s.BaseURL,
		s.InstanceID,
		s.APIKey,
	)

	log.Printf("🔍 Verificando status em: %s", endpoint)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Printf("❌ Erro ao verificar status da instância: %v", err)
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("📥 Resposta completa do status: %s", string(body))
	
	var statusResp ZAPIStatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		log.Printf("❌ Erro ao decodificar resposta do status: %v", err)
		return false, err
	}

	isConnected := statusResp.Connected || statusResp.Smartphon.Connected
	
	log.Printf("📱 Status Z-API - Connected: %v, Session: %s, Phone: %s", 
		isConnected, 
		statusResp.Session,
		statusResp.Smartphon.Number,
	)
	
	if !isConnected {
		log.Printf("⚠️ ATENÇÃO: Instância não está conectada!")
		return false, fmt.Errorf("instância não conectada")
	}

	log.Printf("✅ Instância conectada e pronta!")
	return true, nil
}

// SendMessage envia mensagem via WhatsApp com validações
func (s *WhatsAppService) SendMessage(phone, message string) error {
	// Limpar e formatar número de telefone
	cleanPhone := s.cleanPhoneNumber(phone)
	log.Printf("📞 Enviando para: %s (original: %s)", cleanPhone, phone)

	// Validar número
	if !s.isValidPhone(cleanPhone) {
		log.Printf("❌ Número de telefone inválido: %s", cleanPhone)
		return fmt.Errorf("número de telefone inválido: %s", cleanPhone)
	}

	// Montar URL da API
	endpoint := fmt.Sprintf("%s/instances/%s/token/%s/send-text",
		s.BaseURL,
		s.InstanceID,
		s.APIKey,
	)

	// Preparar payload
	payload := map[string]interface{}{
		"phone":   cleanPhone,
		"message": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ Erro ao serializar payload: %v", err)
		return err
	}

	log.Printf("📤 Endpoint: %s", endpoint)
	log.Printf("📋 Payload: %s", string(jsonData))

	// Fazer requisição
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Erro na requisição HTTP: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Ler resposta completa
	body, _ := io.ReadAll(resp.Body)
	log.Printf("📥 Resposta Z-API [Status: %d]: %s", resp.StatusCode, string(body))

	// Decodificar resposta
	var apiResp ZAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Printf("⚠️ Não foi possível decodificar resposta JSON: %v", err)
		// Continua mesmo sem decodificar, pois pode ser sucesso
	}

	// Verificar erros explícitos na resposta
	if apiResp.Error != "" {
		log.Printf("❌ Erro retornado pela Z-API: %s", apiResp.Error)
		return fmt.Errorf("erro Z-API: %s", apiResp.Error)
	}

	// Verificar status HTTP
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != 201 {
		log.Printf("❌ Z-API retornou status inesperado: %d", resp.StatusCode)
		return fmt.Errorf("erro ao enviar mensagem: status %d - %s", resp.StatusCode, string(body))
	}

	// Verificar se recebemos um ID de mensagem (indicativo de sucesso)
	if apiResp.MessageId != "" || apiResp.ZaapId != "" || apiResp.Id != "" {
		log.Printf("✅ Mensagem enviada com sucesso! ID: %s", 
			firstNonEmpty(apiResp.MessageId, apiResp.ZaapId, apiResp.Id))
		return nil
	}

	log.Printf("✅ Mensagem processada com status %d", resp.StatusCode)
	return nil
}

// firstNonEmpty retorna a primeira string não vazia
func firstNonEmpty(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return "N/A"
}

// SendMessageWithImage envia mensagem com imagem
func (s *WhatsAppService) SendMessageWithImage(phone, message, imageURL string) error {
	cleanPhone := s.cleanPhoneNumber(phone)

	endpoint := fmt.Sprintf("%s/instances/%s/token/%s/send-image",
		s.BaseURL,
		s.InstanceID,
		s.APIKey,
	)

	payload := map[string]interface{}{
		"phone":   cleanPhone,
		"image":   imageURL,
		"caption": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	log.Printf("📤 Enviando imagem para: %s", cleanPhone)

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Erro ao enviar imagem: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("📥 Resposta envio imagem [Status: %d]: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != 201 {
		return fmt.Errorf("erro ao enviar imagem: status %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ Imagem enviada com sucesso para: %s", cleanPhone)
	return nil
}

// cleanPhoneNumber limpa e formata o número de telefone
func (s *WhatsAppService) cleanPhoneNumber(phone string) string {
	// Remove caracteres especiais
	cleaned := strings.NewReplacer(
		"(", "",
		")", "",
		"-", "",
		" ", "",
		"+", "",
		".", "",
	).Replace(phone)

	// Garante que tem o código do país (Brasil: 55)
	if !strings.HasPrefix(cleaned, "55") {
		cleaned = "55" + cleaned
	}

	// Para números brasileiros: 55 + DDD (2 dígitos) + Número (8 ou 9 dígitos)
	// Formato esperado: 5534912345678 (13 dígitos) ou 553491234567 (12 dígitos)
	
	return cleaned
}

// isValidPhone valida o formato do número de telefone brasileiro
func (s *WhatsAppService) isValidPhone(phone string) bool {
	// Número brasileiro deve ter 12 (fixo) ou 13 (celular) dígitos com DDI 55
	if !strings.HasPrefix(phone, "55") {
		return false
	}

	length := len(phone)
	
	// 55 + DDD (2) + número (8 ou 9) = 12 ou 13 dígitos
	if length != 12 && length != 13 {
		log.Printf("⚠️ Tamanho inválido: %d dígitos (esperado: 12 ou 13)", length)
		return false
	}

	// Verificar se são apenas números
	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// TestConnection testa a conexão com a API
func (s *WhatsAppService) TestConnection() error {
	connected, err := s.CheckConnection()
	if err != nil {
		return err
	}
	
	if !connected {
		return fmt.Errorf("instância não está conectada")
	}
	
	log.Printf("✅ Conexão com Z-API OK! Instância pronta para enviar mensagens.")
	return nil
}