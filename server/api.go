package server

import (

	//router and middleware libraries.  Ultimately need to integrate SSO with oauth

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/30x/haystack/oauth2"
	"github.com/30x/haystack/storage"
	"github.com/gorilla/mux"
)

//OAuth libraries

//BasePath the base path all apis extend from
const basePath = "/api"

//TODO make an env variable.  1G max
const maxFileSize = 1024 * 1024 * 1024

//CreateRoutes create a new base api route
func CreateRoutes(storage storage.Storage, authService oauth2.OAuthService) *mux.Router {

	//create our wrapper to point to the storage impl
	api := &API{
		storage:     storage,
		authService: authService,
	}

	r := mux.NewRouter().PathPrefix(basePath).Subrouter()

	r.Path("/bundles").Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").HandlerFunc(api.PostBundle)

	r.Path("/bundles/{bundleName}/revisions").Methods("GET").HandlerFunc(api.GetRevisions)

	r.Path("/bundles/{bundleName}/revisions/{revision}").Methods("GET").HandlerFunc(api.GetBundleRevision)

	r.Path("/bundles/{bundleName}/tags").Methods("POST").HandlerFunc(api.CreateTag)
	r.Path("/bundles/{bundleName}/tags").Methods("GET").HandlerFunc(api.GetTags)

	r.Path("/bundles/{bundleName}/tags/{tagName}").Methods("GET").HandlerFunc(api.GetTag)
	r.Path("/bundles/{bundleName}/tags/{tagName}").Methods("DELETE").HandlerFunc(api.GetTag)

	return r
}

//PostBundle post a bundle
func (a *API) PostBundle(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxFileSize)

	if err != nil {
		writeErrorResponse(http.StatusBadRequest, fmt.Sprintf("Your file cannot be larger than %d bytes", maxFileSize), w)
		return
	}

	bundleName, ok := getBundleName(r.Form)

	if !ok {
		writeErrorResponse(http.StatusBadRequest, "You must specifiy the bundleName parameter", w)
		return
	}

	file, _, err := r.FormFile("bundleData")

	if err != nil {
		writeErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Unable to upload file %s", err), w)
		return
	}

	defer file.Close()

	sha, err := a.storage.SaveBundle(file, bundleName)

	if err != nil {
		writeErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Unable to upload bundle %s", err), w)
		return
	}

	//TODO, not sure this is the best way to render the URL.  Review the http package in more detail and figure out something better before launch
	scheme := r.URL.Scheme

	if scheme == "" {
		scheme = "http"
	}

	response := &BundleCreatedResponse{
		Revision: sha,
		Self:     fmt.Sprintf("%s://%s/api/bundles/%s/revisions/%s", scheme, r.Host, bundleName, sha),
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		writeErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Unable to serialize response  %s", err), w)
	}
}

//GetRevisions get revisions for a bundle
func (a *API) GetRevisions(w http.ResponseWriter, r *http.Request) {
	params := parseBundleRequest(r)

	errs := params.Validate()

	if errs.HasErrors() {
		writeErrorResponses(http.StatusBadRequest, errs, w)
		return
	}

	_, _, err := a.storage.GetRevisions(params.bundleName, "", 100)

	if err != nil {
		writeErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}

	//loop through and recreate the revisions response

}

//GetBundleRevision get bundle data for the revision
func (a *API) GetBundleRevision(w http.ResponseWriter, r *http.Request) {
	params := parseRevisionRequest(r)

	errs := params.Validate()

	if errs.HasErrors() {
		writeErrorResponses(http.StatusBadRequest, errs, w)
		return
	}

	dataReader, err := a.storage.GetBundle(params.bundleName, params.revision)

	if err == storage.ErrRevisionNotExist {
		writeErrorResponse(http.StatusNotFound, fmt.Sprintf("Could not find bundle with name '%s' and revision '%s'", params.bundleName, params.revision), w)
		return
	}

	if err != nil {
		writeErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Could not retrieve bundle. %s", err), w)
		return
	}

	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, dataReader)

	if err != nil {
		writeErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Could not retrieve bundle. %s", err), w)
		return
	}

}

//CreateTag delete the bundle revision
func (a *API) CreateTag(w http.ResponseWriter, r *http.Request) {

}

//GetTags delete the bundle revision
func (a *API) GetTags(w http.ResponseWriter, r *http.Request) {

}

//GetTag delete the bundle revision
func (a *API) GetTag(w http.ResponseWriter, r *http.Request) {

}

//DeleteTag delete the bundle revision
func (a *API) DeleteTag(w http.ResponseWriter, r *http.Request) {

}

//The API instance with the storage pointer
type API struct {
	storage     storage.Storage
	authService oauth2.OAuthService
}

//write a non 200 error response
func writeErrorResponse(statusCode int, message string, w http.ResponseWriter) {

	errors := Errors{message}

	writeErrorResponses(statusCode, errors, w)
}

func writeErrorResponses(statusCode int, errors Errors, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(errors)
}

func getBundleName(formValues url.Values) (string, bool) {
	vals, ok := formValues["bundleName"]

	if !ok || len(vals) != 1 {
		return "", false
	}

	return vals[0], true

}

//a request that required revision and bundle name in the url
type revisionRequest struct {
	bundleRequest
	revision string
}

func parseRevisionRequest(r *http.Request) *revisionRequest {

	vars := mux.Vars(r)

	revisionRequest := &revisionRequest{}

	bundleName, ok := vars["bundleName"]

	if ok {
		revisionRequest.bundleName = bundleName
	}

	revision, ok := vars["revision"]

	if ok {
		revisionRequest.revision = revision
	}

	return revisionRequest
}

func (r *revisionRequest) Validate() Errors {

	errors := Errors{}

	if r.bundleName == "" {
		errors = append(errors, "You must specify a bundle name")
	}

	if r.revision == "" {
		errors = append(errors, "You must specify a revision")
	}

	return errors

}

//a request that required bundle name in the url
type bundleRequest struct {
	bundleName string
}

func parseBundleRequest(r *http.Request) *bundleRequest {

	vars := mux.Vars(r)

	bundleRequest := &bundleRequest{}

	bundleName, ok := vars["bundleName"]

	if ok {
		bundleRequest.bundleName = bundleName
	}

	return bundleRequest
}

func (r *bundleRequest) Validate() Errors {

	errors := Errors{}

	if r.bundleName == "" {
		errors = append(errors, "You must specify a bundle name")
	}

	return errors

}
