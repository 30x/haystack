package storage

import "io"

//Storage the interface for bundle storage
type Storage interface {

	//StoreBundle store the bytes of the bundle id
	SaveBundle(bytes io.Reader, id string) error

	//Get the bundle and return it
	GetBundle(id string) (io.ReadCloser, error)
}
