package response

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, data any) {
	write(w, statusCode, Envelope{
		Code:    statusCode,
		Message: "ok",
		Data:    data,
	})
}

func BadRequest(w http.ResponseWriter, message string) {
	write(w, http.StatusBadRequest, Envelope{
		Code:    http.StatusBadRequest,
		Message: message,
	})
}

func Unauthorized(w http.ResponseWriter, message string) {
	write(w, http.StatusUnauthorized, Envelope{
		Code:    http.StatusUnauthorized,
		Message: message,
	})
}

func NotFound(w http.ResponseWriter, message string) {
	write(w, http.StatusNotFound, Envelope{
		Code:    http.StatusNotFound,
		Message: message,
	})
}

func MethodNotAllowed(w http.ResponseWriter) {
	write(w, http.StatusMethodNotAllowed, Envelope{
		Code:    http.StatusMethodNotAllowed,
		Message: "method not allowed",
	})
}

func write(w http.ResponseWriter, statusCode int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
