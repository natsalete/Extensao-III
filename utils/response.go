package utils

import (
	"encoding/json"
	"net/http"
)

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, response JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func WriteSuccessJSON(w http.ResponseWriter, data interface{}, message string) {
	WriteJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func WriteErrorJSON(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, JSONResponse{
		Success: false,
		Error:   message,
	})
}