package storage

import (
	"errors"
	"io"
	"time"
)

//Storage the interface for bundle storage
type Storage interface {

	//SaveBundle store the bytes of the bundle id.  Returns the new revision and any error
	SaveBundle(bytes io.Reader, owner *BundleMeta) (string, error)

	//GetBundle get the bundle and return it
	GetBundle(bundleMeta *BundleMeta, revision string) (io.ReadCloser, error)

	//GetRevisions get the revisions for the bundle and return them.
	GetRevisions(bundleMeta *BundleMeta, cursor string, pageSize int) ([]*Revision, string, error)

	//CreateTag create a tag for the bundle id. Will return ErrRevisionNotExist if the revision does not exist
	CreateTag(bundleMeta *BundleMeta, revision, tag string) error

	//GetTags get the tags for the bundle. TODO, maybe make this an iterator for the return?
	GetTags(bundleMeta *BundleMeta, cursor string, pageSize int) ([]*Tag, string, error)

	//GetRevisionForTag Get the revision of the bundle and tag.  If none is specified an error will be returned
	GetRevisionForTag(bundleMeta *BundleMeta, tag string) (string, error)

	//DeleteTag a tag for the bundleId and tag.  If the tag does not exist, a ErrTagNotExist will be reteurned
	DeleteTag(bundleMeta *BundleMeta, tag string) error
}

var (
	//ErrRevisionNotExist returned when a revision does not exist
	ErrRevisionNotExist = errors.New("Revision in bucket does not exist")

	//ErrTagNotExist returned when a tag does not exist
	ErrTagNotExist = errors.New("Requested tag in bundle does not exist")

	//ErrNotAllowed The user is not allowed to access this bundle
	ErrNotAllowed = errors.New("The user is not allowed to access this bundle")
)

//Tag a structure to return names and revisions of tags
type Tag struct {
	//The bundle name
	BundleID string

	//RevisionSha512 the sha revision of the tag
	RevisionSha512 string

	//The name of the tag
	Name string

	//the timestamp the tag was created
	Created time.Time
}

//Revision when a revision is created
type Revision struct {
	//The bundle name
	BundleID string

	//The SHa512 revision of the bundle
	RevisionSha512 string

	//the timestamp the bundle was created
	Created time.Time
}

//BundleMeta the owner of the bundle
type BundleMeta struct {
	//The bundle name
	BundleID string
	//The creator's userID from their JWT token
	OwnerUserID string
}
