package storage_test

import (
	"bytes"
	"io/ioutil"

	"github.com/30x/haystack/storage"
	t "github.com/30x/haystack/test"
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
			data := t.CreateFakeBinary(1024)

			bundleId := uuid.NewV1().String()

			sha, err := storageImpl.SaveBundle(bytes.NewReader(data), bundleId)

			t.BeNil(err)

			expectedSha := t.DoSha(data)

			Expect(sha).Should(Equal(expectedSha))

			//now retrieve it

			bundleData, err := storageImpl.GetBundle(bundleId, sha)

			t.BeNil(err)

			returnedBytes, err := ioutil.ReadAll(bundleData)

			t.BeNil(err)

			Expect(returnedBytes).Should(Equal(data))

		})

		It("Get Bundle Revisions", func() {

			data1 := t.CreateFakeBinary(100)
			data2 := t.CreateFakeBinary(101)

			bundleId := uuid.NewV1().String()

			sha1, err := storageImpl.SaveBundle(bytes.NewReader(data1), bundleId)

			t.BeNil(err)

			expectedSha := t.DoSha(data1)

			Expect(sha1).Should(Equal(expectedSha))

			sha2, err := storageImpl.SaveBundle(bytes.NewReader(data2), bundleId)

			t.BeNil(err)

			expectedSha = t.DoSha(data2)

			Expect(sha2).Should(Equal(expectedSha))

			//now retrieve it

			result, err := storageImpl.GetRevisions(bundleId)

			t.BeNil(err)

			Expect(len(result)).Should(Equal(2))

			Expect(result[0]).Should(Equal(sha1))
			Expect(result[1]).Should(Equal(sha2))
		})

		It("Missing bundle Get", func() {
			bundleId := "bundlethatshouldntexist"
			sha := "bad sha"

			reader, err := storageImpl.GetBundle(bundleId, sha)

			t.BeNil(reader)

			Expect(err).Should(Equal(storage.ErrRevisionNotExist))

		})

		It("Create get and list tags", func() {
			//save a 1 k file and then create a tag for it

			bundleId := uuid.NewV1().String()

			data1 := t.CreateFakeBinary(1024)

			sha1, err := storageImpl.SaveBundle(bytes.NewReader(data1), bundleId)

			//simulates a new rev
			t.BeNil(err)

			data2 := t.CreateFakeBinary(20)

			sha2, err := storageImpl.SaveBundle(bytes.NewReader(data2), bundleId)

			t.BeNil(err)

			firstTag := "tag1"
			secondTag := "tag2"
			thirdTag := "tag3"

			err = storageImpl.CreateTag(bundleId, sha1, firstTag)

			t.BeNil(err)

			err = storageImpl.CreateTag(bundleId, sha1, secondTag)

			t.BeNil(err)

			err = storageImpl.CreateTag(bundleId, sha2, thirdTag)

			t.BeNil(err)

			revision, err := storageImpl.GetRevisionForTag(bundleId, firstTag)

			t.BeNil(err)

			Expect(revision).Should(Equal(sha1))

			revision, err = storageImpl.GetRevisionForTag(bundleId, secondTag)

			t.BeNil(err)

			Expect(revision).Should(Equal(sha1))

			revision, err = storageImpl.GetRevisionForTag(bundleId, thirdTag)

			t.BeNil(err)

			Expect(revision).Should(Equal(sha2))

			tags, err := storageImpl.GetTags(bundleId)

			t.BeNil(err)

			Expect(len(tags)).Should(Equal(3))

			Expect(tags[0].Name).Should(Equal(firstTag))
			Expect(tags[0].Revision).Should(Equal(sha1))

			Expect(tags[1].Name).Should(Equal(secondTag))
			Expect(tags[1].Revision).Should(Equal(sha1))

			Expect(tags[2].Name).Should(Equal(thirdTag))
			Expect(tags[2].Revision).Should(Equal(sha2))
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

			t.BeEmpty(sha)

			Expect(err).Should(Equal(storage.ErrTagNotExist))
		})
	}

	//Set up and execute the gcloud implementation for the tests.   Other implementations will define a new context with it's own setup, and execute the tests
	Context("GCloud storage", func() {

		var bucketName string

		BeforeSuite(func() {

			bucketName, storageImpl = t.CreateGCloudImpl()

		})

		AfterSuite(func() {
			t.RemoveGCloudTestBucket(bucketName, storageImpl)
		})

		TestStorage()
	})

})
