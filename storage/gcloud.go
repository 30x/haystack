package storage

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

type gCloudStorageImpl struct {
	bucket *storage.BucketHandle
}

//CreateGCloudStorage create the s3 storage provider and return it.
func CreateGCloudStorage(projectID, bucketName string) (Storage, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)

	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(bucketName)

	_, err = bucket.Attrs(ctx)

	if err != nil {
		// not something we can handle, bail
		if err != storage.ErrObjectNotExist {
			return nil, err
		}

		//try and create on init
		if err := bucket.Create(ctx, projectID, nil); err != nil {
			return nil, err
		}
	}

	// Creates the new bucket

	return &gCloudStorageImpl{
		bucket: bucket,
	}, nil
}

//StoreBundle store the bytes of the bundle id
func (s *gCloudStorageImpl) SaveBundle(bytes io.Reader, id string) (string, error) {

	return "", nil
}

//Get the bundle and return it
func (s *gCloudStorageImpl) GetBundle(id, revision string) (io.ReadCloser, error) {
	return nil, nil
}

//Create a tag for the bundle id
func (s *gCloudStorageImpl) CreateTag(bundleID, revision, tag string) error {
	return nil
}

//GetRevisionForTag Get the revision of the bundle and tag.  If none is specified an error will be returned
func (s *gCloudStorageImpl) GetRevisionForTag(bundleID, tag string) (string, error) {
	return "", nil
}

//DeleteTag a tag for the bundleId and tag.  If the tag does not exist, and error will be reteurned
func (s *gCloudStorageImpl) DeleteTag(bundleID, tag string) error {
	return nil
}
