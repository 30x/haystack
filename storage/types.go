package storage

import "io"

//Storage the interface for bundle storage
type Storage interface {

	//StoreBundle store the bytes of the bundle id.  Returns the new revision and any error
	SaveBundle(bytes io.Reader, id string) (string, error)

	//Get the bundle and return it
	GetBundle(id, revision string) (io.ReadCloser, error)
}
