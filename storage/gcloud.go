package storage

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

type gCloudStorageImpl struct {
	client string
}

//CreateGCloudStorage create the s3 storage provider and return it.
func CreateGCloudStorage(bucketName string) (Storage, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &gCloudStorageImpl{
		client: client,
	}, nil
}

//StoreBundle store the bytes of the bundle id
func (s *s3StorageImpl) SaveBundle(bytes io.Reader, id string) (string, error) {

	return "", nil
}

//Get the bundle and return it
func (s *s3StorageImpl) GetBundle(id, revision string) (io.ReadCloser, error) {
	return nil, nil
}
