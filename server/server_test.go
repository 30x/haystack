package server_test

import (
	"net/http/httptest"

	"github.com/30x/haystack/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("server", func() {

	var testServer string
	TestApi := func() {

		It("Invalid Bundle Id", func() {
		})

	}

	//Set up and execute the gcloud implementation for the tests.   Other implementations will define a new context with it's own setup, and execute the tests
	Context("GCloud storage", func() {

		var bucketName string
		var storageImpl storage.Storage

		BeforeSuite(func() {

			bucketName, storageImpl = storage_test.CreateGCloudImpl()
			r := api.CreateBaseRoute()

			testServer = httptest.NewServer(r)

		})

		AfterSuite(func() {
			storage_test.RemoveGCloudTestBucket(bucketName, storageImpl)
		})

		TestStorage()
	})

})

func beNil(obj interface{}) {
	Expect(obj).Should(BeNil())
}

func beEmpty(obj string) {
	Expect(obj).Should(BeEmpty())
}
