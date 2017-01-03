package api

import (
	"time"

	"github.com/30x/haystack/httputil"
)

//BundleCreatedResponse the created response for the api
type BundleCreatedResponse struct {
	Revision string `json:"revision"`
	Self     string `json:"self"`
}

//BundleRevisions the revisions of bundles
type BundleRevisions struct {
	collection
	Revisions []*RevisionEntry `json:"revisions"`
}

//RevisionEntry the revision entry
type RevisionEntry struct {
	//Revision the revision of this entry
	BundleCreatedResponse
	//The date this revision was stored.
	Created time.Time `json:"date"`
}

//Collection a base type for collections
type collection struct {
	Self   string `json:"self"`
	Cursor string `json:"cursor"`
}

//TagCreate The input payload for the tag create
type TagCreate struct {
	Revision string `json:"revision"`
	Tag      string `json:"tag"`
}

//TagInfo a response of the tag creation
type TagInfo struct {
	TagCreate
	Self string `json:"self"`
}

//TagsResponse the revisions of bundles
type TagsResponse struct {
	collection
	Tags []*TagInfo `json:"tags"`
}

//Validate perform validation on the input
func (t *TagCreate) Validate() httputil.Errors {
	var errors httputil.Errors

	if t.Revision == "" {
		errors = append(errors, "You must specify a revision parammeter")
	}

	if t.Tag == "" {
		errors = append(errors, "You must specify a tag parammeter")
	}

	return errors

}
