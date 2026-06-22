package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, code int, payload interface{}) {
	jsonData, err := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	_, _ = w.Write(jsonData)
}

func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, map[string]string{"error": message})
}
