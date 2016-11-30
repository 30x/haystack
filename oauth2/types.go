package oauth2

import(
  "net/http"
)
//OAuthService the service for verifying OAuth keys
type apigeeOAuth struct{

}

//VerifyOAuth verify the oAuth tokens and permissions
func (a *apigeeOAuth) VerifyOAuth(next http.Handler) http.Handler{
  return nil
}
