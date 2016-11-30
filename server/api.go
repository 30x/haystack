package server

import (

	//router and middleware libraries.  Ultimately need to integrate SSO with oauth

	"net/http"

	"github.com/30x/haystack/storage"
	"github.com/gorilla/mux"
)

//OAuth libraries

//BasePath the base path all apis extend from
const basePath = "/api"

//CreateRoutes create a new base api route
func CreateRoutes(storage storage.Storage) *mux.Router {

	//create our wrapper to point to the storage impl
	api := &Api{
		storage: storage,
	}

	r := mux.NewRouter().PathPrefix(basePath).Subrouter()

	r.Path("/bundles").Methods("POST").HandlerFunc(api.PostBundle)

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
func (a *Api) PostBundle(w http.ResponseWriter, r *http.Request) {

}

//GetRevisions get revisions for a bundle
func (a *Api) GetRevisions(w http.ResponseWriter, r *http.Request) {

}

//GetBundleRevision get bundle data for the revision
func (a *Api) GetBundleRevision(w http.ResponseWriter, r *http.Request) {

}

//DeleteBundleRevision delete the bundle revision
func (a *Api) DeleteBundleRevision(w http.ResponseWriter, r *http.Request) {

}

//CreateTag delete the bundle revision
func (a *Api) CreateTag(w http.ResponseWriter, r *http.Request) {

}

//GetTags delete the bundle revision
func (a *Api) GetTags(w http.ResponseWriter, r *http.Request) {

}

//GetTag delete the bundle revision
func (a *Api) GetTag(w http.ResponseWriter, r *http.Request) {

}

//DeleteTag delete the bundle revision
func (a *Api) DeleteTag(w http.ResponseWriter, r *http.Request) {

}

//The Api instance with the storage pointer
type Api struct {
	storage storage.Storage
}
