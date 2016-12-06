package server

import "time"

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
	collection
	Revisions []*RevisionEntry `json:"revisions"`
}

//RevisionEntry the revision entry
type RevisionEntry struct {
	//Revision the revision of this entry
	Revision string `json:"revision"`
	Self     string `json:"self"`
	//The date this revision was stored.
	Created time.Time `json:"date"`
}

//Collection a base type for collections
type collection struct {
	Self   string `json:"self"`
	Cursor string `json:"cursor"`
}
