package httpserver

import (
	"encoding/json"
	"net/http"
	"time"
)

type ErrorBody struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

type errorBody = ErrorBody

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	WriteJSON(w, status, map[string]any{"error": ErrorBody{
		Code: code, Message: message, RequestID: RequestID(r), Timestamp: time.Now().UTC(),
	}})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	WriteJSON(w, status, body)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	WriteError(w, r, status, code, message)
}

