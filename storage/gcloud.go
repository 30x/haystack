package storage

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/api/iterator"

	uuid "github.com/satori/go.uuid"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
)

//GCloudStorageImpl  The google cloud storage implementation
type GCloudStorageImpl struct {
	Bucket   *storage.BucketHandle
	DsClient *datastore.Client
	Context  context.Context
}

//CreateGCloudStorage create the s3 storage provider and return it.  The serviceAccountFile can be empty, in which case defaults are used.
//see https://cloud.google.com/vision/docs/common/auth for setting creds
func CreateGCloudStorage(projectID, bucketName string) (Storage, error) {

	ctx := context.Background()

	client, err := storage.NewClient(ctx)

	if err != nil {
		return nil, err
	}

	dsClient, err := datastore.NewClient(ctx, projectID)

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

		log.Printf("Creating bucket %s in project %s", bucketName, projectID)
		//try and create on init
		if err := bucket.Create(ctx, projectID, nil); err != nil {
			return nil, err
		}
	}

	// Creates the new bucket

	return &GCloudStorageImpl{
		Bucket:   bucket,
		Context:  ctx,
		DsClient: dsClient,
	}, nil
}

//SaveBundle store the bytes of the bundle id
func (s *GCloudStorageImpl) SaveBundle(bytes io.Reader, bundleMeta *BundleMeta) (string, error) {

	if bundleMeta.BundleID == "" {
		return "", errors.New("You must specify a bundle id")
	}

	timestamp := time.Now()

	tempObjectName := getTempUploadPath(bundleMeta.BundleID)

	tempObject := s.Bucket.Object(tempObjectName)

	writer := tempObject.NewWriter(s.Context)

	//mark the type as a zip before we upload
	writer.ContentType = "application/zip"

	//get the bundle meta, and ensure the owners are the same

	//we have to do get+ write for the first time in a transation to ensure we don't have a race condition
	_, err := s.DsClient.RunInTransaction(s.Context, func(transaction *datastore.Transaction) error {

		existing := &BundleMeta{}

		metaKey := createBundleMetaKey(bundleMeta.BundleID)

		err := transaction.Get(metaKey, existing)

		if err != nil {
			//entity doesn't exist, create it
			if err == datastore.ErrNoSuchEntity {
				_, err := transaction.Put(metaKey, bundleMeta)

				return err

			}
			//if we got it, check they're the same
		} else if bundleMeta.OwnerUserID != existing.OwnerUserID {
			return ErrNotAllowed
		}

		return nil

	})

	if err != nil {
		return "", err
	}

	// io.Copy(writer, bytes)

	//tee the upload so we can calculate the sha

	shaReader := io.TeeReader(bytes, writer)

	hasher := sha512.New()

	log.Printf("Copying bytes for bundleId %s to gcloud and sha512 sum ", bundleMeta.BundleID)

	//reading from the shaReader will also cause the bytes to be copied to the writer, which is in turn sending them to gcloud
	size, err := io.Copy(hasher, shaReader)

	log.Printf("Finished copying %d bytes for bundleId %s", size, bundleMeta.BundleID)

	if err != nil {
		return "", err
	}

	err = writer.Close()

	if err != nil {
		return "", err
	}

	sha512 := hex.EncodeToString(hasher.Sum(nil))

	//now rename to the target file
	targetFile := getRevisionData(bundleMeta.BundleID, sha512)

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

	//write the revision into the cloud db
	revision := &Revision{
		BundleID:       bundleMeta.BundleID,
		RevisionSha512: sha512,
		Created:        timestamp,
	}

	//create hte key and write it.
	key := createRevisionKey(bundleMeta.BundleID, sha512)

	_, err = s.DsClient.Put(s.Context, key, revision)

	return sha512, err
}

//GetBundle the bundle and return it
func (s *GCloudStorageImpl) GetBundle(bundleMeta *BundleMeta, sha512 string) (io.ReadCloser, error) {

	err := s.checkAccess(bundleMeta)

	if err != nil {
		return nil, err
	}

	targetFile := getRevisionData(bundleMeta.BundleID, sha512)

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

//GetBundleOwner get a bundle's owner
func (s *GCloudStorageImpl) GetBundleOwner(bundleID string) (*BundleMeta, error) {
	return nil, nil
}

//GetRevisions get the revisions for the bundle and return them.
func (s *GCloudStorageImpl) GetRevisions(bundleMeta *BundleMeta, cursor string, pageSize int) ([]*Revision, string, error) {

	err := s.checkAccess(bundleMeta)

	if err != nil {
		return nil, "", err
	}

	query := datastore.NewQuery(typeRevision).Namespace(namespace).Limit(pageSize).Ancestor(createBundleMetaKey(bundleMeta.BundleID)).Order("-Created")

	//set the cursor if passed
	if cursor != "" {
		cursor, err := datastore.DecodeCursor(cursor)

		if err != nil {
			return nil, "", err
		}

		query = query.Start(cursor)
	}

	itrResults := s.DsClient.Run(s.Context, query)

	revisions := []*Revision{}

	for {

		revision := &Revision{}

		_, err := itrResults.Next(revision)

		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, "", err
		}

		revisions = append(revisions, revision)
	}

	returnedCursor, err := itrResults.Cursor()

	if err != nil {
		return nil, "", err
	}

	returnCursor := ""

	if len(revisions) == pageSize {
		returnCursor = returnedCursor.String()
	}

	return revisions, returnCursor, nil
}

//CreateTag create a tag for the bundle id
func (s *GCloudStorageImpl) CreateTag(bundleMeta *BundleMeta, sha512, tag string) error {

	err := s.checkAccess(bundleMeta)

	if err != nil {
		return err
	}

	//check if the tag already exists
	revisionKey := createRevisionKey(bundleMeta.BundleID, sha512)

	revision := &Revision{}

	err = s.DsClient.Get(s.Context, revisionKey, revision)

	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return ErrRevisionNotExist
		}

		return err
	}

	key := createTagKey(bundleMeta.BundleID, tag)

	//ensure we get a not found, otherwise we want to bail
	tagData := &Tag{
		Created:        time.Now().UTC(),
		Name:           tag,
		RevisionSha512: sha512,
		BundleID:       bundleMeta.BundleID,
	}

	//write the tag data

	_, err = s.DsClient.Put(s.Context, key, tagData)

	return err

}

//GetTags get the tags
func (s *GCloudStorageImpl) GetTags(bundleMeta *BundleMeta, cursor string, pageSize int) ([]*Tag, string, error) {

	err := s.checkAccess(bundleMeta)

	if err != nil {
		return nil, "", err
	}

	query := datastore.NewQuery(typeTag).Namespace(namespace).Limit(pageSize).Ancestor(createBundleMetaKey(bundleMeta.BundleID)).Order("-Created")

	// query = query.Filter("BundleID = ", bundleMeta.BundleID).Filter("OwnerUserID = ", bundleMeta.OwnerUserID)

	//set the cursor if passed
	if cursor != "" {
		cursor, err := datastore.DecodeCursor(cursor)

		if err != nil {
			return nil, "", err
		}

		query = query.Start(cursor)
	}

	itrResults := s.DsClient.Run(s.Context, query)

	tags := []*Tag{}

	for {

		tag := &Tag{}

		_, err := itrResults.Next(tag)

		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, "", err
		}

		tags = append(tags, tag)
	}

	returnedCursor, err := itrResults.Cursor()

	if err != nil {
		return nil, "", err
	}

	returnCursor := ""

	if len(tags) == pageSize {
		returnCursor = returnedCursor.String()
	}

	return tags, returnCursor, nil

}

//GetRevisionForTag Get the revision of the bundle and tag.  If none is specified an error will be returned
func (s *GCloudStorageImpl) GetRevisionForTag(bundleMeta *BundleMeta, tag string) (string, error) {

	err := s.checkAccess(bundleMeta)

	if err != nil {
		return "", err
	}

	tagEntity := &Tag{}

	err = s.DsClient.Get(s.Context, createTagKey(bundleMeta.BundleID, tag), tagEntity)

	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return "", ErrTagNotExist
		}

		return "", err

	}

	return tagEntity.RevisionSha512, nil

}

//checkAccess check if the requested user has access
func (s *GCloudStorageImpl) checkAccess(requestedBundleMeta *BundleMeta) error {

	existingMeta := &BundleMeta{}

	err := s.DsClient.Get(s.Context, createBundleMetaKey(requestedBundleMeta.BundleID), existingMeta)

	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return ErrRevisionNotExist
		}

		return err
	}

	if requestedBundleMeta.OwnerUserID != existingMeta.OwnerUserID {
		return ErrNotAllowed
	}

	return nil
}

//DeleteTag a tag for the bundleId and tag.  If the tag does not exist, and error will be reteurned
func (s *GCloudStorageImpl) DeleteTag(bundleMeta *BundleMeta, tag string) error {

	//make sure it exists
	_, err := s.GetRevisionForTag(bundleMeta, tag)

	if err != nil {
		return err
	}

	key := createTagKey(bundleMeta.BundleID, tag)

	return s.DsClient.Delete(s.Context, key)
}

func getTempUploadPath(bundleID string) string {
	return fmt.Sprintf("%s/uploading/%s", bundleID, uuid.NewV1().String())
}

func getRevisionData(bundleID, revision string) string {
	return fmt.Sprintf("%s/revisionData/%s.zip", bundleID, revision)
}

func createRevisionKey(bundleID, revision string) *datastore.Key {
	return &datastore.Key{
		Parent:    createBundleMetaKey(bundleID),
		Name:      fmt.Sprintf("%s-sha512:%s", bundleID, revision),
		Kind:      typeRevision,
		Namespace: namespace,
	}

}

func createBundleMetaKey(bundleID string) *datastore.Key {
	return &datastore.Key{
		Name:      bundleID,
		Kind:      typeBundleMeta,
		Namespace: namespace,
	}

}

func createTagKey(bundleID, tag string) *datastore.Key {
	return &datastore.Key{
		Parent:    createBundleMetaKey(bundleID),
		Name:      fmt.Sprintf("%s-%s", bundleID, tag),
		Kind:      typeTag,
		Namespace: namespace,
	}

}

const typeRevision = "Revision"
const typeBundleMeta = "BundleMeta"
const typeTag = "Tag"
const namespace = "BundleStorage"
