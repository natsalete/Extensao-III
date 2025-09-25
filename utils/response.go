package utils

import (
	"encoding/json"
	"net/http"
)

// Estrutura única que cobre sucesso e erro
type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Função genérica para escrever JSON
func WriteJSON(w http.ResponseWriter, status int, response JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

// Resposta de sucesso (com ou sem dados)
func SendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	WriteJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Resposta de erro (com status customizável)
func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	WriteJSON(w, statusCode, JSONResponse{
		Success: false,
		Error:   message,
	})
}
