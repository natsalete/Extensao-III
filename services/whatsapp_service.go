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
	ClientToken string
	InstanceID  string
	BaseURL     string
}

// Estruturas de resposta
type ZAPIStatusResponse struct {
	Connected bool   `json:"connected"`
	Session   string `json:"session"`
	Error     string `json:"error"`
	Smartphon struct {
		Connected bool   `json:"connected"`
		Number    string `json:"number"`
	} `json:"smartphon"`
}

type ZAPIResponse struct {
	ZaapId    string `json:"zaapId"`
	MessageId string `json:"messageId"`
	Id        string `json:"id"`
	Error     string `json:"error"`
	Message   string `json:"message"`
}

// Inicializa o serviço com variáveis de ambiente
func NewWhatsAppService() *WhatsAppService {
	return &WhatsAppService{
		APIKey:      os.Getenv("WHATSAPP_API_KEY"),
		ClientToken: os.Getenv("WHATSAPP_CLIENT_TOKEN"),
		InstanceID:  os.Getenv("WHATSAPP_INSTANCE_ID"),
		BaseURL:     "https://api.z-api.io",
	}
}

// Verifica se a instância está conectada
func (s *WhatsAppService) CheckConnection() (bool, error) {
	endpoint := fmt.Sprintf("%s/instances/%s/token/%s/status/%s",
		s.BaseURL,
		s.InstanceID,
		s.APIKey,
		s.ClientToken,
	)

	log.Printf("🔍 Verificando status em: %s", endpoint)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Printf("❌ Erro ao verificar status: %v", err)
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("📥 Resposta status: %s", string(body))

	var status ZAPIStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		log.Printf("❌ Erro ao decodificar resposta JSON: %v", err)
		return false, err
	}

	if !status.Connected && !status.Smartphon.Connected {
		return false, fmt.Errorf("⚠️ Instância não conectada")
	}

	log.Printf("✅ Instância conectada! Sessão: %s, Número: %s",
		status.Session, status.Smartphon.Number)
	return true, nil
}

// Envia uma mensagem de texto
func (s *WhatsAppService) SendMessage(phone, message string) error {
	cleanPhone := s.cleanPhoneNumber(phone)

	if !s.isValidPhone(cleanPhone) {
		return fmt.Errorf("número de telefone inválido: %s", cleanPhone)
	}

	// ✅ URL correta
	endpoint := fmt.Sprintf("%s/instances/%s/token/%s/send-text",
		s.BaseURL,
		s.InstanceID,
		s.APIKey,
	)

	payload := map[string]interface{}{
		"phone":   cleanPhone,
		"message": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", s.ClientToken) // ✅ NOVO: Token via cabeçalho

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("📤 Resposta envio: %s", string(body))

	var result ZAPIResponse
	json.Unmarshal(body, &result)

	if result.Error != "" {
		return fmt.Errorf("erro Z-API: %s", result.Error)
	}

	log.Printf("✅ Mensagem enviada com sucesso! ID: %s",
		firstNonEmpty(result.MessageId, result.ZaapId, result.Id))

	return nil
}


// Envia uma imagem com legenda
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

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Token", s.ClientToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("📷 Resposta envio imagem: %s", string(body))
	return nil
}


// Funções auxiliares
func (s *WhatsAppService) cleanPhoneNumber(phone string) string {
	cleaned := strings.NewReplacer("(", "", ")", "", "-", "", " ", "", "+", "", ".", "").Replace(phone)
	if !strings.HasPrefix(cleaned, "55") {
		cleaned = "55" + cleaned
	}
	return cleaned
}

func (s *WhatsAppService) isValidPhone(phone string) bool {
	length := len(phone)
	if length != 12 && length != 13 {
		return false
	}
	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func firstNonEmpty(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return "N/A"
}

func (s *WhatsAppService) TestConnection() error {
	connected, err := s.CheckConnection()
	if err != nil {
		return err
	}
	if !connected {
		return fmt.Errorf("instância não está conectada")
	}
	log.Printf("✅ Conexão com Z-API OK!")
	return nil
}
