package storage_test

import (
	"time"

	"github.com/30x/haystack/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("storage", func() {

	var storage storage.Storage

	TestStorage := func() {
		PIt("Invalid Storage Input", func() {

		})

		PIt("Invalid Sha Format", func() {

		})

		PIt("Empty reader", func() {

		})

		PIt("Valid Bundle Save + GET", func() {

		})

		PIt("Missing bundle Get", func() {

		})

		PIt("Create and get tag", func() {

		})

		PIt("Create tag missing revision", func() {

		})

		PIt("Delete tag missing revision", func() {

		})

		PIt("Get tag missing tag", func() {

		})
	}

	//Set up and execute the gcloud implementation for the tests.   Other implementations will define a new context with it's own setup, and execute the tests
	Context("GCloud storage", func() {

		var bucketName string

		BeforeEach(func() {

			bucketName = time.Now().Unix()

			storage, error = CreateGCloudStorage(bucketName)

			Expect(error).Should(BeNil(), "Could not create local docker image creator")

		})

		TestStorage()
	})

})
