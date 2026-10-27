package httpserver

import (
	"encoding/json"
	"net/http"
	"time"
)

type errorBody struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": errorBody{
		Code: code, Message: message, RequestID: requestID(r), Timestamp: time.Now().UTC(),
	}})
}
