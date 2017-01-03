package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SermoDigital/jose/jwt"
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

const tokenKey = "ssotoken"

//SetTokenInRequest sets a token into the request
func SetTokenInRequest(r *http.Request, token jwt.JWT) *http.Request {
	newContext := context.WithValue(r.Context(), tokenKey, token)

	return r.WithContext(newContext)
}

//GetTokenFromRequest get the SSO token from the request
func GetTokenFromRequest(r *http.Request) (jwt.JWT, error) {
	value := r.Context().Value(tokenKey)

	//shouldn't happen at runtime, indicates an error
	if value == nil {
		return nil, errors.New("No token was set into the request")
	}

	return value.(jwt.JWT), nil
}
