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
	. "github.com/30x/haystack/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("server", func() {

	var testServer *httptest.Server

	TestApi := func() {
		It("Bundle Upload and Get", func() {
			testPayload := CreateFakeBinary(1024)

			url := testServer.URL + "/api/bundles"

			bundleName := "test"

			response, body, err := newFileUploadRequest(url, bundleName, testPayload)

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(200))

			bundleCreatedResponse := &server.BundleCreatedResponse{}

			err = json.Unmarshal([]byte(body), bundleCreatedResponse)

			IsNil(err)

			expectedSha := DoSha(testPayload)

			Expect(bundleCreatedResponse.Revision).Should(Equal(expectedSha))

			expectedUrl := fmt.Sprintf("%s/api/bundles/%s/revisions/%s", testServer.URL, bundleName, expectedSha)

			Expect(bundleCreatedResponse.Self).Should(Equal(expectedUrl))

			//now perform the get and assert they're equal
			var fileBytes []byte

			response, errors := submitGetRequest(bundleCreatedResponse.Self, func(body []byte) {
				fileBytes = body
			})

			IsNil(errors)

			Expect(response.StatusCode).Should(Equal(200))

			Expect(fileBytes).Should(Equal(testPayload))

		})
	}

	//Set up and execute the gcloud implementation for the tests.   Other implementations will define a new context with it's own setup, and execute the tests
	Context("GCloud storage", func() {

		var bucketName string
		var storageImpl storage.Storage

		BeforeSuite(func() {

			bucketName, storageImpl = CreateGCloudImpl()
			fakeOauth := &noOpTestAuth{}
			r := server.CreateRoutes(storageImpl, fakeOauth)

			testServer = httptest.NewServer(r)

		})

		AfterSuite(func() {
			RemoveGCloudTestBucket(bucketName, storageImpl)
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

//submit the get request.  If the response is 200, then the parser will be invoked.  Otherwise errors will be parsed
func submitGetRequest(url string, parser func(body []byte)) (*http.Response, *server.Errors) {

	request, err := http.NewRequest("GET", url, nil)

	IsNil(err)

	client := &http.Client{}

	response, err := client.Do(request)

	IsNil(err)

	defer response.Body.Close()

	responseBody := resposneBodyAsBytes(response)

	// fmt.Printf("Response body is %s", string(responseBody))

	var errors *server.Errors

	//only parse our org if it's a successfull response code
	if response.StatusCode == 200 {
		parser(responseBody)

	} else if len(responseBody) > 0 {
		//ignore error on error unmarshall.
		json.Unmarshal(responseBody, &errors)

	}

	return response, errors
}

//responseBodyAsBytes get the response body and close it properly
func resposneBodyAsBytes(response *http.Response) []byte {
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	IsNil(err)

	return bodyBytes

}

type noOpTestAuth struct{}

func (n *noOpTestAuth) VerifyOAuth(next http.Handler) http.Handler{
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(rw, r)
		})
}
