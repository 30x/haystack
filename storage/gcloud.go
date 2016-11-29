package storage

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"google.golang.org/api/iterator"

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

	timestamp := time.Now()

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
	targetFile := getRevisionData(bundleID, sha512)

	destinationObject := s.Bucket.Object(targetFile)

	_, err = destinationObject.CopierFrom(tempObject).Run(s.Context)

	if err != nil {
		return "", err
	}

	//now delete the original
	err = tempObject.Delete(s.Context)

	if err != nil {
		return "", err
	}

	orderedRevision := getRevisionPath(bundleID, sha512, timestamp)

	writer = s.Bucket.Object(orderedRevision).NewWriter(s.Context)

	_, err = writer.Write([]byte{0})

	if err != nil {
		return "", err
	}

	err = writer.Close()

	if err != nil {
		return "", err
	}

	return sha512, nil
}

//GetBundle the bundle and return it
func (s *GCloudStorageImpl) GetBundle(bundleID, sha512 string) (io.ReadCloser, error) {
	targetFile := getRevisionData(bundleID, sha512)

	destinationObject := s.Bucket.Object(targetFile)

	reader, err := destinationObject.NewReader(s.Context)

	if err != nil {

		if err == storage.ErrObjectNotExist {
			return nil, ErrRevisionNotExist
		}

		return nil, err
	}

	return reader, nil
}

//GetRevisions get the revisions for the bundle and return them.
func (s *GCloudStorageImpl) GetRevisions(bundleID string) ([]string, error) {

	revisions := []string{}

	//scan all tags for the bundleid
	itr := s.Bucket.Objects(s.Context, &storage.Query{
		Prefix: fmt.Sprintf("%s/revisions", bundleID),
	})

	for {
		obj, err := itr.Next()

		if err != nil {
			if err == iterator.Done {
				break
			}

			return nil, err
		}

		parts := strings.Split(obj.Name, "-")

		revision := parts[len(parts)-1]

		revisions = append(revisions, revision)

	}

	return revisions, nil
}

//CreateTag create a tag for the bundle id
func (s *GCloudStorageImpl) CreateTag(bundleID, sha512, tag string) error {

	targetFile := getRevisionData(bundleID, sha512)

	//check it exists
	_, err := s.Bucket.Object(targetFile).Attrs(s.Context)

	if err != nil {
		if err == storage.ErrObjectNotExist {
			return ErrRevisionNotExist
		}

		return err
	}

	//now create the file

	tagPath := getTagPath(bundleID, tag)

	tagObject := s.Bucket.Object(tagPath)

	writer := tagObject.NewWriter(s.Context)

	_, err = io.Copy(writer, strings.NewReader(sha512))

	if err != nil {
		return err
	}

	//close the output to the file
	err = writer.Close()

	return err
}

//GetTags get the tags
func (s *GCloudStorageImpl) GetTags(bundleID string) ([]*Tag, error) {

	tags := []*Tag{}

	//scan all tags for the bundleid
	itr := s.Bucket.Objects(s.Context, &storage.Query{
		Prefix: fmt.Sprintf("%s/tags", bundleID),
	})

	for {
		obj, err := itr.Next()

		if err != nil {
			if err == iterator.Done {
				break
			}

			return nil, err
		}

		parts := strings.Split(obj.Name, "/")

		tagName := parts[len(parts)-1]

		//TODO make this fan out/merge for faster execution with lots of tags
		shaValue, err := s.getShaFromTag(obj.Name)

		if err != nil {
			return nil, err
		}

		tags = append(tags, &Tag{
			Name:     tagName,
			Revision: shaValue,
		})

	}

	return tags, nil

}

//GetRevisionForTag Get the revision of the bundle and tag.  If none is specified an error will be returned
func (s *GCloudStorageImpl) GetRevisionForTag(bundleID, tag string) (string, error) {
	tagPath := getTagPath(bundleID, tag)

	return s.getShaFromTag(tagPath)
}

func (s *GCloudStorageImpl) getShaFromTag(tagPath string) (string, error) {
	reader, err := s.Bucket.Object(tagPath).NewReader(s.Context)

	if err != nil {
		if err == storage.ErrObjectNotExist {
			return "", ErrTagNotExist
		}
		return "", err
	}

	shaBuffer := &bytes.Buffer{}

	_, err = io.Copy(shaBuffer, reader)

	if err != nil {
		return "", err
	}

	//now we've copied return the string
	return shaBuffer.String(), nil
}

//DeleteTag a tag for the bundleId and tag.  If the tag does not exist, and error will be reteurned
func (s *GCloudStorageImpl) DeleteTag(bundleID, tag string) error {

	tagPath := getTagPath(bundleID, tag)

	err := s.Bucket.Object(tagPath).Delete(s.Context)

	if err == storage.ErrObjectNotExist {
		return ErrTagNotExist
	}

	return err
}

func getTempUploadPath(bundleID string) string {
	return fmt.Sprintf("%s/uploading/%s", bundleID, uuid.NewV1().String())
}

func getRevisionData(bundleID, revision string) string {
	return fmt.Sprintf("%s/revisionData/%s.zip", bundleID, revision)
}

//get the path where a revision pointer is stored
func getRevisionPath(bundleID, revision string, timestamp time.Time) string {
	//take the timestamp and minus the max so we get reverse ordering

	orderID := timestamp.UTC().Format(time.RFC3339Nano)

	// return fmt.Sprintf("%s/revisions/%020d-%s", bundleID, orderID, revision)
	return fmt.Sprintf("%s/revisions/%s-%s", bundleID, orderID, revision)
}

func getTagPath(bundleID, tag string) string {
	return fmt.Sprintf("%s/tags/%s", bundleID, tag)
}
