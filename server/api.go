package server

import (

	//router and middleware libraries.  Ultimately need to integrate SSO with oauth

	"net/http"

	"github.com/gorilla/mux"
)

//OAuth libraries

//BasePath the base path all apis extend from
const basePath = "/api"

//CreateRoutes create a new base api route
func CreateRoutes() *mux.Router {

	r := mux.NewRouter().PathPrefix(basePath).Subrouter()

	r.Path("/bundles").Methods("POST").HandlerFunc(PostBundle)

	r.Path("/bundles/{bundleName}/revisions").Methods("GET").HandlerFunc(GetRevisions)

	r.Path("/bundles/{bundleName}/revisions/{revision}").Methods("GET").HandlerFunc(GetBundleRevision)

	r.Path("/bundles/{bundleName}/revisions/{revision}").Methods("DELETE").HandlerFunc(DeleteBundleRevision)

	r.Path("/bundles/{bundleName}/tags").Methods("POST").HandlerFunc(CreateTag)
	r.Path("/bundles/{bundleName}/tags").Methods("GET").HandlerFunc(GetTags)

	r.Path("/bundles/{bundleName}/tags/{tagName}").Methods("GET").HandlerFunc(GetTag)
	r.Path("/bundles/{bundleName}/tags/{tagName}").Methods("DELETE").HandlerFunc(GetTag)

	return r

}

//PostBundle post a bundle
func PostBundle(w http.ResponseWriter, r *http.Request) {

}

//GetRevisions get revisions for a bundle
func GetRevisions(w http.ResponseWriter, r *http.Request) {

}

//GetBundleRevision get bundle data for the revision
func GetBundleRevision(w http.ResponseWriter, r *http.Request) {

}

//DeleteBundleRevision delete the bundle revision
func DeleteBundleRevision(w http.ResponseWriter, r *http.Request) {

}

//CreateTag delete the bundle revision
func CreateTag(w http.ResponseWriter, r *http.Request) {

}

//GetTags delete the bundle revision
func GetTags(w http.ResponseWriter, r *http.Request) {

}

//GetTag delete the bundle revision
func GetTag(w http.ResponseWriter, r *http.Request) {

}

//DeleteTag delete the bundle revision
func DeleteTag(w http.ResponseWriter, r *http.Request) {

}
