package storage_test

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/30x/haystack/storage"
	. "github.com/30x/haystack/test"
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
			Expect(sha).Should(BeEmpty())
			Expect(err.Error()).Should(Equal("You must specify a bundle id"))
		})

		It("Empty reader", func() {
			data := [...]byte{}
			sha, err := storageImpl.SaveBundle(bytes.NewReader(data[:len(data)]), "")
			Expect(sha).Should(BeEmpty())
			Expect(err.Error()).Should(Equal("You must specify a bundle id"))
		})

		It("Valid Bundle Save + GET", func() {

			//1k
			data := CreateFakeBinary(1024)

			bundleId := uuid.NewV1().String()

			sha, err := storageImpl.SaveBundle(bytes.NewReader(data), bundleId)

			IsNil(err)

			expectedSha := DoSha(data)

			Expect(sha).Should(Equal(expectedSha))

			//now retrieve it

			bundleData, err := storageImpl.GetBundle(bundleId, sha)

			IsNil(err)

			returnedBytes, err := ioutil.ReadAll(bundleData)

			IsNil(err)

			Expect(returnedBytes).Should(Equal(data))

		})

		It("Get Bundle Revisions", func() {

			//tests bundle revisions with paging
			size := uint32(5)

			savedShas := make([]string, size)

			bundleId := uuid.NewV1().String()

			writeStarted := time.Now()

			for i := uint32(0); i < size; i++ {

				fileData := GenerateBinaryFromInt(i)

				sha, err := storageImpl.SaveBundle(bytes.NewReader(fileData), bundleId)

				IsNil(err)

				expectedSha := DoSha(fileData)

				Expect(sha).Should(Equal(expectedSha))

				savedShas[i] = sha

			}

			//now retrieve and test

			result, cursor, err := storageImpl.GetRevisions(bundleId, "", 2)

			IsNil(err)
			Expect(cursor).ShouldNot(BeEmpty())

			Expect(len(result)).Should(Equal(2))

			Expect(result[0].Revision).Should(Equal(savedShas[0]))
			Expect(writeStarted.Before(result[0].Created)).Should(BeTrue())

			Expect(result[1].Revision).Should(Equal(savedShas[1]))
			Expect(writeStarted.Before(result[1].Created)).Should(BeTrue())

			result, cursor, err = storageImpl.GetRevisions(bundleId, cursor, 2)

			IsNil(err)
			Expect(cursor).ShouldNot(BeEmpty())

			Expect(len(result)).Should(Equal(2))
			Expect(result[0].Revision).Should(Equal(savedShas[2]))
			Expect(writeStarted.Before(result[0].Created)).Should(BeTrue())

			Expect(result[1].Revision).Should(Equal(savedShas[3]))
			Expect(writeStarted.Before(result[1].Created)).Should(BeTrue())

			result, cursor, err = storageImpl.GetRevisions(bundleId, cursor, 2)

			IsNil(err)

			Expect(cursor).Should(BeEmpty())

			Expect(len(result)).Should(Equal(1))

			Expect(result[0].Revision).Should(Equal(savedShas[4]))
			Expect(writeStarted.Before(result[0].Created)).Should(BeTrue())

		})

		It("Missing bundle Get", func() {
			bundleId := "bundlethatshouldntexist"
			sha := "bad sha"

			reader, err := storageImpl.GetBundle(bundleId, sha)

			IsNil(reader)

			Expect(err).Should(Equal(storage.ErrRevisionNotExist))

		})

		It("Create get and list tags", func() {
			//save a 1 k file and then create a tag for it

			bundleId := uuid.NewV1().String()

			data1 := CreateFakeBinary(1024)

			sha1, err := storageImpl.SaveBundle(bytes.NewReader(data1), bundleId)

			//simulates a new rev
			IsNil(err)

			data2 := CreateFakeBinary(20)

			sha2, err := storageImpl.SaveBundle(bytes.NewReader(data2), bundleId)

			IsNil(err)

			firstTag := "tag1"
			secondTag := "tag2"
			thirdTag := "tag3"

			err = storageImpl.CreateTag(bundleId, sha1, firstTag)

			IsNil(err)

			err = storageImpl.CreateTag(bundleId, sha1, secondTag)

			IsNil(err)

			err = storageImpl.CreateTag(bundleId, sha2, thirdTag)

			IsNil(err)

			revision, err := storageImpl.GetRevisionForTag(bundleId, firstTag)

			IsNil(err)

			Expect(revision).Should(Equal(sha1))

			revision, err = storageImpl.GetRevisionForTag(bundleId, secondTag)

			IsNil(err)

			Expect(revision).Should(Equal(sha1))

			revision, err = storageImpl.GetRevisionForTag(bundleId, thirdTag)

			IsNil(err)

			Expect(revision).Should(Equal(sha2))

			tags, _, err := storageImpl.GetTags(bundleId, "", 100)

			IsNil(err)

			Expect(len(tags)).Should(Equal(3))

			Expect(tags[0].Name).Should(Equal(firstTag))
			Expect(tags[0].Revision).Should(Equal(sha1))

			Expect(tags[1].Name).Should(Equal(secondTag))
			Expect(tags[1].Revision).Should(Equal(sha1))

			Expect(tags[2].Name).Should(Equal(thirdTag))
			Expect(tags[2].Revision).Should(Equal(sha2))
		})

		It("List tags", func() {

			//tests bundle revisions with paging
			bundleId := uuid.NewV1().String()

			data1 := CreateFakeBinary(10)

			sha1, err := storageImpl.SaveBundle(bytes.NewReader(data1), bundleId)

			//simulates a new rev
			IsNil(err)

			data2 := CreateFakeBinary(11)

			sha2, err := storageImpl.SaveBundle(bytes.NewReader(data2), bundleId)

			IsNil(err)

			//now create 5 tags, first 2 on sha1, second 2 on sha 2 last one on sha 1 and iterate through them

			tag1 := "tag1"
			tag2 := "tag2"
			tag3 := "tag3"
			tag4 := "tag4"
			tag5 := "tag5"

			err = storageImpl.CreateTag(bundleId, sha1, tag1)

			IsNil(err)

			err = storageImpl.CreateTag(bundleId, sha1, tag2)

			IsNil(err)

			err = storageImpl.CreateTag(bundleId, sha2, tag3)

			IsNil(err)

			err = storageImpl.CreateTag(bundleId, sha2, tag4)

			IsNil(err)

			err = storageImpl.CreateTag(bundleId, sha1, tag5)

			IsNil(err)

			result, cursor, err := storageImpl.GetTags(bundleId, "", 2)

			IsNil(err)
			Expect(cursor).ShouldNot(BeEmpty())

			Expect(len(result)).Should(Equal(2))

			Expect(result[0].Name).Should(Equal(tag1))
			Expect(result[0].Revision).Should(Equal(sha1))

			Expect(result[1].Name).Should(Equal(tag2))
			Expect(result[1].Revision).Should(Equal(sha1))

			result, cursor, err = storageImpl.GetTags(bundleId, cursor, 2)

			IsNil(err)
			Expect(cursor).ShouldNot(BeEmpty())

			Expect(len(result)).Should(Equal(2))

			Expect(result[0].Name).Should(Equal(tag3))
			Expect(result[0].Revision).Should(Equal(sha2))

			Expect(result[1].Name).Should(Equal(tag4))
			Expect(result[1].Revision).Should(Equal(sha2))

			result, cursor, err = storageImpl.GetTags(bundleId, cursor, 2)

			IsNil(err)

			Expect(len(result)).Should(Equal(1))

			Expect(result[0].Name).Should(Equal(tag5))
			Expect(result[0].Revision).Should(Equal(sha1))

			Expect(cursor).Should(BeEmpty())

		})

		It("Create tag missing revision", func() {

			bundleId := "testbundle"
			revision := "1234"

			//try to create a tag on something that doesn't exist
			err := storageImpl.CreateTag(bundleId, revision, "test")

			Expect(err).Should(Equal(storage.ErrRevisionNotExist))
		})

		It("Delete tag missing", func() {
			bundleId := "testbundle"
			tag := "test"

			//try to create a tag on sometrhing that doesn't exist
			err := storageImpl.DeleteTag(bundleId, tag)

			Expect(err).Should(Equal(storage.ErrTagNotExist))
		})

		It("Get tag missing tag", func() {
			bundleId := "testbundle"
			tag := "test"
			sha, err := storageImpl.GetRevisionForTag(bundleId, tag)

			IsEmpty(sha)

			Expect(err).Should(Equal(storage.ErrTagNotExist))
		})
	}

	//Set up and execute the gcloud implementation for the tests.   Other implementations will define a new context with it's own setup, and execute the tests
	Context("GCloud storage", func() {

		var bucketName string

		BeforeSuite(func() {

			bucketName, storageImpl = CreateGCloudImpl()

		})

		AfterSuite(func() {
			RemoveGCloudTestBucket(bucketName, storageImpl)
		})

		TestStorage()
	})

})
