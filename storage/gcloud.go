package storage

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"

	uuid "github.com/satori/go.uuid"

	"cloud.google.com/go/storage"
)

//GCloudStorageImpl  The google cloud storage implementation
type GCloudStorageImpl struct {
	Bucket  *storage.BucketHandle
	Context context.Context
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
		if err != storage.ErrBucketNotExist {
			return nil, err
		}

		log.Printf("Creating bucket in project %s", projectID)
		//try and create on init
		if err := bucket.Create(ctx, projectID, nil); err != nil {
			return nil, err
		}
	}

	// Creates the new bucket

	return &GCloudStorageImpl{
		Bucket:  bucket,
		Context: ctx,
	}, nil
}

//SaveBundle store the bytes of the bundle id
func (s *GCloudStorageImpl) SaveBundle(bytes io.Reader, bundleID string) (string, error) {

	if bundleID == "" {
		return "", errors.New("You must specify a bundle id")
	}

	tempObjectName := getTempUploadPath(bundleID)

	tempObject := s.Bucket.Object(tempObjectName)

	writer := tempObject.NewWriter(s.Context)

	//mark the type as a zip before we upload
	writer.ContentType = "application/zip"

	// io.Copy(writer, bytes)

	//tee the upload so we can calculate the sha

	shaReader := io.TeeReader(bytes, writer)

	hasher := sha512.New()

	log.Printf("Copying bytes for bundleId %s to gcloud and sha512 sum ", bundleID)

	//reading from the shaReader will also cause the bytes to be copied to the writer, which is in turn sending them to gcloud
	size, err := io.Copy(hasher, shaReader)

	log.Printf("Finished copying %d bytes for bundleId %s", size, bundleID)

	if err != nil {
		return "", err
	}

	err = writer.Close()

	if err != nil {
		return "", err
	}

	sha512 := hex.EncodeToString(hasher.Sum(nil))

	//now rename to the target file
	targetFile := getRevisionPath(bundleID, sha512)

	destinationObject := s.Bucket.Object(targetFile)

	_, err = destinationObject.CopierFrom(tempObject).Run(s.Context)

	if err != nil {
		return "", err
	}

	//now delete the origina
	err = tempObject.Delete(s.Context)

	if err != nil {
		return "", err
	}

	return sha512, nil

	// return "", nil
}

//GetBundle the bundle and return it
func (s *GCloudStorageImpl) GetBundle(id, revision string) (io.ReadCloser, error) {
	return nil, nil
}

//CreateTag create a tag for the bundle id
func (s *GCloudStorageImpl) CreateTag(bundleID, revision, tag string) error {
	return nil
}

//GetRevisionForTag Get the revision of the bundle and tag.  If none is specified an error will be returned
func (s *GCloudStorageImpl) GetRevisionForTag(bundleID, tag string) (string, error) {
	return "", nil
}

//DeleteTag a tag for the bundleId and tag.  If the tag does not exist, and error will be reteurned
func (s *GCloudStorageImpl) DeleteTag(bundleID, tag string) error {
	return nil
}

func getTempUploadPath(bundleID string) string {
	return fmt.Sprintf("%s/uploading/%s", bundleID, uuid.NewV1().String())
}

func getRevisionPath(bundleID, revision string) string {
	return fmt.Sprintf("%s/revisions/%s.zip", bundleID, revision)
}

func getTagsPath(bundleID string) string {
	return fmt.Sprintf("%s/tags/", bundleID)
}

func getTagPath(bundleID, tag string) string {
	return fmt.Sprintf("%s/tags/%s", bundleID, tag)
}
