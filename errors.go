package main

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func sendJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(APIError{
		Error: message,
		Code:  code,
	})
}
