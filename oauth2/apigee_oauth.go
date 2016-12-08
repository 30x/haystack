package oauth2

import (
	"net/http"
)

//OAuthService the service for verifying OAuth keys
type apigeeOAuth struct {
}

//VerifyOAuth verify the oAuth tokens and permissions
func (a *apigeeOAuth) VerifyOAuth(next http.Handler) http.Handler {
	return nil
}

//CreateApigeeOAuth create an apigee instance of the oauth service
func CreateApigeeOAuth() OAuthService {
	return &apigeeOAuth{}
}
