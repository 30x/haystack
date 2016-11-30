package server

//BundleCreatedResponse the created response for the api
type BundleCreatedResponse struct {
	Revision string `json:"revision"`
	Self     string `json:"self"`
}

//Errors to return
type Errors []string

//HasErrors return true if there are errors
func (e Errors) HasErrors() bool {
	return len(e) > 0
}

//BundleRevisions the revisions of bundles
type BundleRevisions struct {
}

//RevisionEntry the revision entry
type RevisionEntry struct {
}

//Collection a base type for collections
type Collection struct {
	Self   string `json:"self"`
	Cursor string `json:"cursor"`
}
