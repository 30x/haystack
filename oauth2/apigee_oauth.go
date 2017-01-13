package oauth2

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/30x/haystack/httputil"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
)

//OAuthService the service for verifying OAuth keys
type apigeeOAuth struct {
	keyURL string
	client *http.Client
}

//VerifyOAuth verify the oAuth tokens and permissions
func (a *apigeeOAuth) VerifyOAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		jwt, err := jws.ParseJWTFromRequest(r)

		if err != nil {
			httputil.WriteErrorResponse(http.StatusBadRequest, err.Error(), rw)
			return
		}

		err = a.Validate(jwt)

		if err != nil {
			httputil.WriteErrorResponse(http.StatusBadRequest, err.Error(), rw)
			return
		}

		//set our JWT into the request, it's valid
		newRequest := SetPrincipalInRequest(r, &apigeePrincipal{jwtToken: jwt})

		next.ServeHTTP(rw, newRequest)
	})

}

//ValidateKey validate the jwt and return an error if it fails
func (a *apigeeOAuth) Validate(jwt jwt.JWT) error {

	//TODO move this into a cache
	r, err := a.client.Get(a.keyURL)

	if err != nil {
		return err
	}

	defer r.Body.Close()

	ssoKey := &ssoKey{}

	err = json.NewDecoder(r.Body).Decode(ssoKey)

	if err != nil {
		return err
	}

	//now validate to the token.

	publieKey, err := crypto.ParseRSAPublicKeyFromPEM([]byte(ssoKey.Value))

	if err != nil {
		return err
	}

	return jwt.Validate(publieKey, crypto.SigningMethodRS256)
}

//CreateApigeeOAuth create an apigee instance of the oauth service
func CreateApigeeOAuth(keyURL string) OAuthService {
	return &apigeeOAuth{
		keyURL: keyURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type ssoKey struct {
	Alg   string `json:"alg"`
	Value string `json:"value"`
	Kty   string `json:"kty"`
	Use   string `json:"use"`
	N     string `json:"n"`
	E     string `json:"e"`
}

type apigeePrincipal struct {
	jwtToken jwt.JWT
}

func (a *apigeePrincipal) GetSubject() (string, error) {
	subject, exists := a.jwtToken.Claims().Subject()

	if !exists {
		return "", errors.New("Cannot get subject from JWT token")
	}

	return subject, nil
}
