package oauth2

import(
  "net/http"
)
//OAuthService the service for verifying OAuth keys
type OAuthService interface{
  //VerifyOAuth verify the oAuth tokens and permissions
  VerifyOAuth(next http.Handler) http.Handler
}
