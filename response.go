package chi

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON body with the given status code.
// Note: WriteHeader is called before Encode, so once JSON starts writing
// you can no longer change the status. Validate/marshal-check earlier if
// that matters for a given handler.
func JSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// Error writes a JSON error envelope: {"error": "..."}.
func Error(w http.ResponseWriter, status int, message string) error {
	return JSON(w, status, map[string]string{"error": message})
}

// Success writes a JSON data envelope: {"data": ...}.
func Success(w http.ResponseWriter, status int, data any) error {
	return JSON(w, status, map[string]any{"data": data})
}
