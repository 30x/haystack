package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/30x/haystack/server"
	"github.com/30x/haystack/storage"
	"github.com/30x/haystack/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("server", func() {

	var testServer *httptest.Server

	TestApi := func() {
		It("Good Bundle Payload", func() {
			payload := test.CreateFakeBinary(1024)

			url := testServer.URL + "/api/bundles"

			bundleName := "test"

			response, body, err := newFileUploadRequest(url, bundleName, payload)

			test.BeNil(err)

			Expect(response.StatusCode).Should(Equal(200))

			bundleCreatedResponse := &server.BundleCreatedResponse{}

			err = json.Unmarshal([]byte(body), bundleCreatedResponse)

			test.BeNil(err)

			expectedSha := test.DoSha(payload)

			Expect(bundleCreatedResponse.Revision).Should(Equal(expectedSha))

			expectedUrl := fmt.Sprintf("%s/bundles/%s/revisions/%s", testServer.URL, bundleName, expectedSha)

			Expect(bundleCreatedResponse.Self).Should(Equal(expectedUrl))
		})
	}

	//Set up and execute the gcloud implementation for the tests.   Other implementations will define a new context with it's own setup, and execute the tests
	Context("GCloud storage", func() {

		var bucketName string
		var storageImpl storage.Storage

		BeforeSuite(func() {

			bucketName, storageImpl = test.CreateGCloudImpl()
			r := server.CreateRoutes(storageImpl)

			testServer = httptest.NewServer(r)

		})

		AfterSuite(func() {
			test.RemoveGCloudTestBucket(bucketName, storageImpl)
		})

		TestApi()
	})

})

func newFileUploadRequest(url string, bundleName string, file []byte) (*http.Response, string, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("bundleData", "data.zip")

	if err != nil {
		return nil, "", err
	}
	//copy over the bytes
	_, err = io.Copy(part, bytes.NewReader(file))

	if err != nil {
		return nil, "", err
	}

	writer.WriteField("bundleName", bundleName)

	//set the content type
	writer.FormDataContentType()

	err = writer.Close()

	if err != nil {
		return nil, "", err
	}

	request, err := http.NewRequest("POST", url, body)

	if err != nil {
		return nil, "", err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	//request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "e30K.e30K.e30K"))

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	response, err := client.Do(request)

	if err != nil {
		return nil, "", err
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, "", err
	}

	bodyResponse := string(bodyBytes)

	return response, bodyResponse, nil
}
