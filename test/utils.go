package test

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"

	gstorage "cloud.google.com/go/storage"

	"github.com/30x/haystack/storage"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/iterator"

	. "github.com/onsi/gomega"
)

//IsNil assert an instance is nil
func IsNil(obj interface{}) {
	Expect(obj).Should(BeNil())
}

//IsEmpty assert a string is empty
func IsEmpty(obj string) {
	Expect(obj).Should(BeEmpty())
}

//CreateFakeBinary create a binary file with the specified number of bytes
func CreateFakeBinary(length int) []byte {

	byteArray := make([]byte, length)

	for i := 0; i < length; i++ {
		byteArray[i] = byte(rand.Intn(255))
	}

	return byteArray
}

//DoSha get the sha512 sum of the bytes provided
func DoSha(data []byte) string {
	hasher := sha512.New()
	hasher.Write(data)

	bytes := hasher.Sum(nil)

	return hex.EncodeToString(bytes)
}

//CreateGCloudImpl returns the bucket name used for the test, and the gcloud storage implementation
func CreateGCloudImpl() (string, storage.Storage) {

	projectID := os.Getenv("PROJECTID")

	Expect(projectID).ShouldNot(BeEmpty(), "You must set the PROJECTID env variable for your gcloud project to run the tests")

	bucketName := "bundle-test-" + uuid.NewV1().String()

	gcloud, err := storage.CreateGCloudStorage(projectID, bucketName)

	Expect(err).Should(BeNil(), "Could not create g cloud storage")

	return bucketName, gcloud

}

//RemoveGCloudTestBucket remove all items in the bucket and it's corresponding bucket when complete
func RemoveGCloudTestBucket(bucketName string, storageImpl storage.Storage) {
	gcloud := (storageImpl.(*storage.GCloudStorageImpl))

	context := context.Background()

	itr := gcloud.Bucket.Objects(context, &gstorage.Query{})

	for {
		obj, err := itr.Next()

		if err == iterator.Done {
			break
		}

		err = gcloud.Bucket.Object(obj.Name).Delete(context)

		Expect(err).Should(BeNil(), fmt.Sprintf("Error when deleting object %s is %s", obj.Name, err))

	}

	//now delete the bucket
	err := gcloud.Bucket.Delete(context)

	Expect(err).Should(BeNil(), "Could not clean up bucket from test")
}
