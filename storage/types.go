package storage

import (
	"errors"
	"io"
	"time"
)

//Storage the interface for bundle storage
type Storage interface {

	//SaveBundle store the bytes of the bundle id.  Returns the new revision and any error
	SaveBundle(bytes io.Reader, bundleID string) (string, error)

	//GetBundle get the bundle and return it
	GetBundle(bundleID, revision string) (io.ReadCloser, error)

	//GetRevisions get the revisions for the bundle and return them.
	GetRevisions(bundleID, cursor string, pageSize int) ([]*Revision, string, error)

	//CreateTag create a tag for the bundle id
	CreateTag(bundleID, revision, tag string) error

	//GetTags get the tags for the bundle. TODO, maybe make this an iterator for the return?
	GetTags(bundleID, cursor string, pageSize int) ([]*Tag, string, error)

	//GetRevisionForTag Get the revision of the bundle and tag.  If none is specified an error will be returned
	GetRevisionForTag(bundleID, tag string) (string, error)

	//DeleteTag a tag for the bundleId and tag.  If the tag does not exist, and error will be reteurned
	DeleteTag(bundleID, tag string) error
}

var (
	//ErrRevisionNotExist returned when a revision does not exist
	ErrRevisionNotExist = errors.New("Revision in bucket does not exist")

	//ErrTagNotExist returned when a tag does not exist
	ErrTagNotExist = errors.New("Tag in bucket does not exist")
)

//Tag a structure to return names and revisions of tags
type Tag struct {
	Revision string
	Name     string
}

//Revision when a revision is created
type Revision struct {
	Revision string
	Created  time.Time
}
