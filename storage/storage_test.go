package storage_test

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"crypto/sha512"

	"github.com/30x/haystack/storage"
	"github.com/satori/go.uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("storage", func() {

	var storageImpl storage.Storage

	TestStorage := func() {

		It("Invalid Bundle Id", func() {
			data := [...]byte{1, 1, 1}
			sha, err := storageImpl.SaveBundle(bytes.NewReader(data[:len(data)]), "")
			Expect(sha).Should(BeNil())
			Expect(err.Error()).Should(Equal("You must specify a bundle id"))
		})

		It("Empty reader", func() {
			data := [...]byte{}
			sha, err := storageImpl.SaveBundle(bytes.NewReader(data[:len(data)]), "")
			Expect(sha).Should(BeNil())
			Expect(err.Error()).Should(Equal("You must specify a bundle id"))
		})

		It("Valid Bundle Save + GET", func() {

			//1k
			data := createFakeBinary(1024)

			bundleId := uuid.NewV1().String()

			sha, err := storageImpl.SaveBundle(bytes.NewReader(data), bundleId)

			beNil(err)

			expectedSha := doSha(data)

			Expect(sha).Should(Equal(expectedSha))

			//now retrieve it

			bundleData, err := storageImpl.GetBundle(bundleId, sha)

			beNil(err)

			returnedBytes, err := ioutil.ReadAll(bundleData)

			beNil(err)

			Expect(returnedBytes).Should(Equal(data))

		})

		It("Missing bundle Get", func() {
			bundleId := "bundlethatshouldntexist"
			sha := "bad sha"

			reader, err := storageImpl.GetBundle(bundleId, sha)

			beNil(reader)

			Expect(err.Error()).Should(Equal(fmt.Sprintf("No bundle with bundleId '%s' and revision '%s' could be found", bundleId, sha)))

		})

		It("Create and get tag", func() {
			//save a 1 k file and then create a tag for it

			bundleId := uuid.NewV1().String()

			data1 := createFakeBinary(1024)

			sha1, err := storageImpl.SaveBundle(bytes.NewReader(data1), bundleId)

			//simulates a new rev
			beNil(err)

			data2 := createFakeBinary(20)

			sha2, err := storageImpl.SaveBundle(bytes.NewReader(data2), bundleId)

			beNil(err)

			firstTag := "tag1"
			secondTag := "tag2"
			thirdTag := "tag3"

			err = storageImpl.CreateTag(bundleId, sha1, firstTag)

			beNil(err)

			err = storageImpl.CreateTag(bundleId, sha1, secondTag)

			beNil(err)

			err = storageImpl.CreateTag(bundleId, sha2, thirdTag)

			beNil(err)

			revision, err := storageImpl.GetRevisionForTag(bundleId, firstTag)

			beNil(err)

			Expect(revision).Should(Equal(sha1))

			revision, err = storageImpl.GetRevisionForTag(bundleId, secondTag)

			beNil(err)

			Expect(revision).Should(Equal(sha1))

			revision, err = storageImpl.GetRevisionForTag(bundleId, thirdTag)

			beNil(err)

			Expect(revision).Should(Equal(sha2))
		})

		It("Create tag missing revision", func() {

			bundleId := "testbundle"
			revision := "1234"

			//try to create a tag on sometrhing that doesn't exist
			err := storageImpl.CreateTag(bundleId, revision, "test")

			Expect(err.Error()).Should(Equal(fmt.Sprintf("No bundle with id '%s' and revision '%s' was found", bundleId, revision)))
		})

		It("Delete tag missing", func() {
			bundleId := "testbundle"
			revision := "1234"
			tag := "test"

			//try to create a tag on sometrhing that doesn't exist
			err := storageImpl.CreateTag(bundleId, revision, tag)

			Expect(err.Error()).Should(Equal(fmt.Sprintf("No tag with name '%s' was found for bundle with id '%s' and revision '%s'", tag, bundleId, revision)))
		})

		It("Get tag missing tag", func() {
			bundleId := "testbundle"
			tag := "test"
			sha, err := storageImpl.GetRevisionForTag(bundleId, tag)

			beNil(sha)

			Expect(err.Error()).Should(Equal(fmt.Sprintf("No tag with name '%s' was found for bundle with id '%s'", tag, bundleId)))
		})
	}

	//Set up and execute the gcloud implementation for the tests.   Other implementations will define a new context with it's own setup, and execute the tests
	Context("GCloud storage", func() {

		var bucketName string

		BeforeEach(func() {

			bucketName = uuid.NewV1().String()

			gcloud, err := storage.CreateGCloudStorage("", bucketName)

			Expect(err).Should(BeNil(), "Could not create g cloud storage")

			storageImpl = gcloud

		})

		TestStorage()
	})

})

func beNil(obj interface{}) {
	Expect(obj).Should(BeNil())
}

func createFakeBinary(length int) []byte {

	byteArray := make([]byte, length)

	for i := 0; i < length; i++ {
		byteArray[i] = 1
	}

	return byteArray
}

func doSha(data []byte) string {
	bytes := sha512.Sum512(data)

	return string(bytes[:len(bytes)])
}
