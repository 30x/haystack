package storage_test

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"

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

		FIt("Get Bundle Revisions", func() {

			//tests bundle revisions with paging
			size := 5

			savedShas := make([]string, size)

			bundleId := uuid.NewV1().String()

			for i := 0; i < size; i++ {

				fileData := generatePayloadFromInt(256)

				sha, err := storageImpl.SaveBundle(bytes.NewReader(fileData), bundleId)

				IsNil(err)

				expectedSha := DoSha(fileData)

				Expect(sha).Should(Equal(expectedSha))

				savedShas[i] = sha

			}

			//now retrieve and test
			pageSize := 2
			iterations := size / pageSize
			cursor := ""
			var err error
			var result []string

			for i := 0; i < iterations; i++ {

				result, cursor, err = storageImpl.GetRevisions(bundleId, cursor, pageSize)

				IsNil(err)

				Expect(cursor).ShouldNot(BeEmpty())

				startIndex := i * pageSize
				length := startIndex + pageSize

				if length > size {
					length = size - 1
				}

				AssertStrings(savedShas, startIndex, length, result)
			}

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

//Check that the strings in expected match every index in sub.  It will start at expected[startIndex] = sub[0] to expected[startIndex+length] = sub[len(sub)]
func AssertStrings(expected []string, startIndex, length int, sub []string) {

	Expect(len(sub)).Should(Equal(startIndex+length), "Sub index length does not match")

	subIndex := 0
	expectedIndex := startIndex
	endIndex := startIndex + length

	Expect(startIndex+length < len(expected)).Should(BeTrue())

	for expectedIndex < endIndex {
		Expect(expected[expectedIndex]).Should(Equal(sub[subIndex]))

		expectedIndex++
		subIndex++
	}

}

func generatePayloadFromInt(index int32) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, index)
	IsNil(err)

	return buf.Bytes()
}
