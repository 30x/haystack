package httputil

import (
	"encoding/json"
	"net/http"
)

//WriteErrorResponse write a non 200 error response
func WriteErrorResponse(statusCode int, message string, w http.ResponseWriter) {

	errors := Errors{message}

	WriteErrorResponses(statusCode, errors, w)
}

//WriteErrorResponses write our error responses
func WriteErrorResponses(statusCode int, errors Errors, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(errors)
}
