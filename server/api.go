package server

import (

	//router and middleware libraries.  Ultimately need to integrate SSO with oauth

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/30x/haystack/storage"
	"github.com/gorilla/mux"
)

//OAuth libraries

//BasePath the base path all apis extend from
const basePath = "/api"

//TODO make an env variable.  1G max
const maxFileSize = 1024 * 1024 * 1024

//CreateRoutes create a new base api route
func CreateRoutes(storage storage.Storage) *mux.Router {

	//create our wrapper to point to the storage impl
	api := &API{
		storage: storage,
	}

	r := mux.NewRouter().PathPrefix(basePath).Subrouter()

	r.Path("/bundles").Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").HandlerFunc(api.PostBundle)

	r.Path("/bundles/{bundleName}/revisions").Methods("GET").HandlerFunc(api.GetRevisions)

	r.Path("/bundles/{bundleName}/revisions/{revision}").Methods("GET").HandlerFunc(api.GetBundleRevision)

	r.Path("/bundles/{bundleName}/revisions/{revision}").Methods("DELETE").HandlerFunc(api.DeleteBundleRevision)

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

	//TODO, not sure this is the best way to render the URL.  Review the http package in more detail
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

}

//GetBundleRevision get bundle data for the revision
func (a *API) GetBundleRevision(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	bundleName, ok := vars["bundleName"]

	if !ok {
		writeErrorResponse(http.StatusBadRequest, "You must specify a bundle name", w)
		return
	}

	revision, ok := vars["revision"]

	if !ok {
		writeErrorResponse(http.StatusBadRequest, "You must specify a revision", w)
		return
	}

	dataReader, err := a.storage.GetBundle(bundleName, revision)

	if err == storage.ErrRevisionNotExist {
		writeErrorResponse(http.StatusNotFound, fmt.Sprintf("Could not find bundle with name '%s' and revision '%s'", bundleName, revision), w)
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

//DeleteBundleRevision delete the bundle revision
func (a *API) DeleteBundleRevision(w http.ResponseWriter, r *http.Request) {

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
	storage storage.Storage
}

//write a non 200 error response
func writeErrorResponse(statusCode int, message string, w http.ResponseWriter) {

	w.WriteHeader(statusCode)

	errors := Errors{message}

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
