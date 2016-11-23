package storage

import "io"

//Storage the interface for bundle storage
type Storage interface {

	//SaveBundle store the bytes of the bundle id.  Returns the new revision and any error
	SaveBundle(bytes io.Reader, bundleID string) (string, error)

	//GetBundle get the bundle and return it
	GetBundle(bundleID, revision string) (io.ReadCloser, error)

	//CreateTag create a tag for the bundle id
	CreateTag(bundleID, revision, tag string) error

	//GetRevisionForTag Get the revision of the bundle and tag.  If none is specified an error will be returned
	GetRevisionForTag(bundleID, tag string) (string, error)

	//DeleteTag a tag for the bundleId and tag.  If the tag does not exist, and error will be reteurned
	DeleteTag(bundleID, tag string) error
}
