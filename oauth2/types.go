package oauth2

import (
	"net/http"
)

//OAuthService the service for verifying OAuth keys
type OAuthService interface {
	//VerifyOAuth verify the oAuth tokens and permissions
	VerifyOAuth(next http.Handler) http.Handler
}

//Principal the principal of the JWT token. This interface primarily exists as a means of decoupling auth during the api testing
type Principal interface {
	//return the identifier of the principal. This should be an immutable value for the principal, otherwise future access to assets is not gaurenteed
	GetSubject() (string, error)
}
