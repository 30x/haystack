package api

import (

	//router and middleware libraries.  Ultimately need to integrate SSO with oauth

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"strconv"

	"github.com/30x/haystack/httputil"
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

	r.Path("/bundles").Methods("POST").HeadersRegexp("Content-Type", "multipart/form-data.*").Handler(authService.VerifyOAuth(http.HandlerFunc(api.PostBundle)))

	r.Path("/bundles/{bundleName}/revisions").Methods("GET").Handler(authService.VerifyOAuth(http.HandlerFunc(api.GetRevisions)))

	r.Path("/bundles/{bundleName}/revisions/{revision}").Methods("GET").Handler(authService.VerifyOAuth(http.HandlerFunc(api.GetBundleRevision)))

	r.Path("/bundles/{bundleName}/tags").Methods("POST").Handler(authService.VerifyOAuth(http.HandlerFunc(api.CreateTag)))
	r.Path("/bundles/{bundleName}/tags").Methods("GET").Handler(authService.VerifyOAuth(http.HandlerFunc(api.GetTags)))

	r.Path("/bundles/{bundleName}/tags/{tagName}").Methods("GET").Handler(authService.VerifyOAuth(http.HandlerFunc(api.GetTag)))
	r.Path("/bundles/{bundleName}/tags/{tagName}").Methods("DELETE").Handler(authService.VerifyOAuth(http.HandlerFunc(api.GetTag)))

	return r
}

//PostBundle post a bundle
func (a *API) PostBundle(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxFileSize)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	bundleName, ok := getBundleName(r.Form)

	if !ok {
		httputil.WriteErrorResponse(http.StatusBadRequest, "You must specifiy the bundleName parameter", w)
		return
	}

	file, _, err := r.FormFile("bundleData")

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Unable to upload file.  Make sure bundleData is present in the form with file data.  Err is %s", err), w)
		return
	}

	defer file.Close()

	principal, err := oauth2.GetPrincipalFromRequest(r)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	subject, err := principal.GetSubject()

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	bundleMeta := &storage.BundleMeta{
		BundleID:    bundleName,
		OwnerUserID: subject,
	}

	sha, err := a.storage.SaveBundle(file, bundleMeta)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Unable to upload bundle %s", err), w)
		return
	}

	//TODO, not sure this is the best way to render the URL.  Review the http package in more detail and figure out something better before launch

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := &BundleCreatedResponse{
		Revision: sha,
		Self:     createRevisionURL(r, bundleName, sha),
	}

	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Unable to serialize response  %s", err), w)
	}
}

//GetRevisions get revisions for a bundle
func (a *API) GetRevisions(w http.ResponseWriter, r *http.Request) {
	params := parseBundleRequest(r)

	errs := params.Validate()

	if errs.HasErrors() {
		httputil.WriteErrorResponses(http.StatusBadRequest, errs, w)
		return
	}

	principal, err := oauth2.GetPrincipalFromRequest(r)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	subject, err := principal.GetSubject()

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	bundleMeta := &storage.BundleMeta{
		BundleID:    params.bundleName,
		OwnerUserID: subject,
	}

	cursor, pageSize, err := parsePaginationValues(r)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	revisions, cursor, err := a.storage.GetRevisions(bundleMeta, cursor, pageSize)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}

	bundleRevisions := &BundleRevisions{}

	bundleRevisions.Cursor = cursor

	for _, savedRevision := range revisions {
		newRev := &RevisionEntry{

			Created: savedRevision.Created,
		}

		newRev.Revision = savedRevision.RevisionSha512
		newRev.Self = createRevisionURL(r, params.bundleName, savedRevision.RevisionSha512)

		bundleRevisions.Revisions = append(bundleRevisions.Revisions, newRev)
	}

	//loop through and recreate the revisions response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bundleRevisions)

}

//GetBundleRevision get bundle data for the revision
func (a *API) GetBundleRevision(w http.ResponseWriter, r *http.Request) {
	params := parseRevisionRequest(r)

	errs := params.Validate()

	if errs.HasErrors() {
		httputil.WriteErrorResponses(http.StatusBadRequest, errs, w)
		return
	}

	principal, err := oauth2.GetPrincipalFromRequest(r)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	subject, err := principal.GetSubject()

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	bundleMeta := &storage.BundleMeta{
		BundleID:    params.bundleName,
		OwnerUserID: subject,
	}

	dataReader, err := a.storage.GetBundle(bundleMeta, params.revision)

	if err == storage.ErrRevisionNotExist {
		httputil.WriteErrorResponse(http.StatusNotFound, fmt.Sprintf("Could not find bundle with name '%s' and revision '%s'", params.bundleName, params.revision), w)
		return
	}

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Could not retrieve bundle. %s", err), w)
		return
	}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, dataReader)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Could not retrieve bundle. %s", err), w)
		return
	}

}

//CreateTag delete the bundle revision
func (a *API) CreateTag(w http.ResponseWriter, r *http.Request) {

	bundleRequest := parseBundleRequest(r)

	errors := bundleRequest.Validate()

	if errors.HasErrors() {
		httputil.WriteErrorResponses(http.StatusBadRequest, errors, w)
		return
	}

	defer r.Body.Close()

	principal, err := oauth2.GetPrincipalFromRequest(r)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	subject, err := principal.GetSubject()

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	bundleMeta := &storage.BundleMeta{
		BundleID:    bundleRequest.bundleName,
		OwnerUserID: subject,
	}

	tagCreate := &TagCreate{}

	err = json.NewDecoder(r.Body).Decode(tagCreate)

	//can't parse the json
	if err != nil {
		httputil.WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Could not parse json. %s", err), w)
		return
	}

	//valid json, but not what we expect
	errors = tagCreate.Validate()

	if errors.HasErrors() {
		httputil.WriteErrorResponses(http.StatusBadRequest, errors, w)
		return
	}

	//create the tags
	err = a.storage.CreateTag(bundleMeta, tagCreate.Revision, tagCreate.Tag)

	if err != nil {
		if err == storage.ErrRevisionNotExist {
			httputil.WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Revision %s does not exist for bundle %s", tagCreate.Revision, bundleRequest.bundleName), w)
			return
		}

		httputil.WriteErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}

	tagInfo := &TagInfo{
		Self: createTagURL(r, bundleRequest.bundleName, tagCreate.Tag),
	}

	tagInfo.Revision = tagCreate.Revision
	tagInfo.Tag = tagCreate.Tag

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(tagInfo)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, err.Error(), w)
	}

}

//GetTags delete the bundle revision
func (a *API) GetTags(w http.ResponseWriter, r *http.Request) {
	params := parseBundleRequest(r)

	errs := params.Validate()

	if errs.HasErrors() {
		httputil.WriteErrorResponses(http.StatusBadRequest, errs, w)
		return
	}

	cursor, pageSize, err := parsePaginationValues(r)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	principal, err := oauth2.GetPrincipalFromRequest(r)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	subject, err := principal.GetSubject()

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	bundleMeta := &storage.BundleMeta{
		BundleID:    params.bundleName,
		OwnerUserID: subject,
	}

	tags, cursor, err := a.storage.GetTags(bundleMeta, cursor, pageSize)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}

	tagsResponse := &TagsResponse{}

	tagsResponse.Cursor = cursor

	for _, savedTag := range tags {
		tagInfo := &TagInfo{
			Self: createTagURL(r, params.bundleName, savedTag.Name),
		}

		tagInfo.Revision = savedTag.RevisionSha512
		tagInfo.Tag = savedTag.Name

		tagsResponse.Tags = append(tagsResponse.Tags, tagInfo)
	}

	//loop through and recreate the revisions response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tagsResponse)
}

//GetTag delete the bundle revision
func (a *API) GetTag(w http.ResponseWriter, r *http.Request) {
	tagRequest := parseTagRequest(r)

	errors := tagRequest.Validate()

	if errors.HasErrors() {
		httputil.WriteErrorResponses(http.StatusBadRequest, errors, w)
		return
	}

	principal, err := oauth2.GetPrincipalFromRequest(r)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	subject, err := principal.GetSubject()

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	bundleMeta := &storage.BundleMeta{
		BundleID:    tagRequest.bundleName,
		OwnerUserID: subject,
	}

	rev, err := a.storage.GetRevisionForTag(bundleMeta, tagRequest.tag)

	if err != nil {
		if err == storage.ErrTagNotExist {
			httputil.WriteErrorResponse(http.StatusNotFound, fmt.Sprintf("Could not find bundle with name '%s' and tag '%s'", tagRequest.bundleName, tagRequest.tag), w)
			return
		}

		httputil.WriteErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}

	//valid, return it
	tagInfo := &TagInfo{
		Self: createTagURL(r, tagRequest.bundleName, tagRequest.tag),
	}

	tagInfo.Revision = rev
	tagInfo.Tag = tagRequest.tag

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(tagInfo)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, err.Error(), w)
	}

}

//DeleteTag delete the bundle revision
func (a *API) DeleteTag(w http.ResponseWriter, r *http.Request) {

	tagRequest := parseTagRequest(r)

	errors := tagRequest.Validate()

	if errors.HasErrors() {
		httputil.WriteErrorResponses(http.StatusBadRequest, errors, w)
		return
	}

	principal, err := oauth2.GetPrincipalFromRequest(r)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	subject, err := principal.GetSubject()

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, "Unable to validate user", w)
		return
	}

	bundleMeta := &storage.BundleMeta{
		BundleID:    tagRequest.bundleName,
		OwnerUserID: subject,
	}

	//now delete it
	rev, err := a.storage.GetRevisionForTag(bundleMeta, tagRequest.tag)

	if err != nil {
		if err == storage.ErrTagNotExist {
			httputil.WriteErrorResponse(http.StatusNotFound, fmt.Sprintf("Could not find bundle with name '%s' and tag '%s'", tagRequest.bundleName, tagRequest.tag), w)
			return
		}

		httputil.WriteErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}

	//now delete it
	err = a.storage.DeleteTag(bundleMeta, tagRequest.tag)

	if err != nil {
		if err == storage.ErrTagNotExist {
			httputil.WriteErrorResponse(http.StatusNotFound, fmt.Sprintf("Could not find bundle with name '%s' and tag '%s'", tagRequest.bundleName, tagRequest.tag), w)
			return
		}

		httputil.WriteErrorResponse(http.StatusInternalServerError, err.Error(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	//valid, return it
	tagInfo := &TagInfo{
		Self: createTagURL(r, tagRequest.bundleName, tagRequest.tag),
	}

	tagInfo.Revision = rev
	tagInfo.Tag = tagRequest.tag

	err = json.NewEncoder(w).Encode(tagInfo)

	if err != nil {
		httputil.WriteErrorResponse(http.StatusInternalServerError, err.Error(), w)
	}
}

//The API instance with the storage pointer
type API struct {
	storage     storage.Storage
	authService oauth2.OAuthService
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

func (r *revisionRequest) Validate() httputil.Errors {

	var errors httputil.Errors

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

func (r *bundleRequest) Validate() httputil.Errors {

	var errors httputil.Errors

	if r.bundleName == "" {
		errors = append(errors, "You must specify a bundle name")
	}

	return errors

}

//Parses pagination values.  Returns the cursor and the page size, if specified.  If not specified a default will be used
func parsePaginationValues(req *http.Request) (string, int, error) {

	values := req.URL.Query()

	cursor := ""

	passedCursor := values.Get("cursor")

	if passedCursor != "" {
		cursor = passedCursor
	}

	pageSize := 100

	passedPageSize := values.Get("pageSize")

	if passedPageSize != "" {
		var err error
		pageSize, err = strconv.Atoi(passedPageSize)

		if err != nil {
			return "", 0, err
		}
	}

	return cursor, pageSize, nil
}

func createRevisionURL(r *http.Request, bundleName, sha string) string {

	scheme := r.URL.Scheme

	if scheme == "" {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s/api/bundles/%s/revisions/%s", scheme, r.Host, bundleName, sha)
}

func createTagURL(r *http.Request, bundleName, tag string) string {

	scheme := r.URL.Scheme

	if scheme == "" {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s/api/bundles/%s/tags/%s", scheme, r.Host, bundleName, tag)
}

//a request that required tag and bundle name in the url
type tagRequest struct {
	bundleRequest
	tag string
}

func parseTagRequest(r *http.Request) *tagRequest {

	vars := mux.Vars(r)

	tagRequest := &tagRequest{}

	bundleName, ok := vars["bundleName"]

	if ok {
		tagRequest.bundleName = bundleName
	}

	tag, ok := vars["tagName"]

	if ok {
		tagRequest.tag = tag
	}

	return tagRequest
}

func (r *tagRequest) Validate() httputil.Errors {

	var errors httputil.Errors

	if r.bundleName == "" {
		errors = append(errors, "You must specify a bundle name")
	}

	if r.tag == "" {
		errors = append(errors, "You must specify a tag")
	}

	return errors

}
