package oauth2

import (
	"context"
	"errors"
	"net/http"
)

const principalKey = "github.com.30x.haystack.principal"

//SetPrincipalInRequest sets a token into the request
func SetPrincipalInRequest(r *http.Request, principal Principal) *http.Request {
	newContext := context.WithValue(r.Context(), principalKey, principal)

	return r.WithContext(newContext)
}

//GetPrincipalFromRequest get the SSO token from the request
func GetPrincipalFromRequest(r *http.Request) (Principal, error) {
	value := r.Context().Value(principalKey)

	//shouldn't happen at runtime, indicates an error
	if value == nil {
		return nil, errors.New("No principal was set into the request")
	}

	return value.(Principal), nil
}
