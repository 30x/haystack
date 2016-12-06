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
			testPayload := CreateFakeBinary(10)

			bundleName := "test"

			response, bundleCreatedResponse, err := uploadBundle(testServer, bundleName, bytes.NewReader(testPayload))

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(http.StatusOK))

			expectedSha := DoSha(testPayload)

			Expect(bundleCreatedResponse.Revision).Should(Equal(expectedSha))

			expectedUrl := fmt.Sprintf("%s/api/bundles/%s/revisions/%s", testServer.URL, bundleName, expectedSha)

			Expect(bundleCreatedResponse.Self).Should(Equal(expectedUrl))

			//now perform the get and assert they're equal
			var fileBytes []byte

			response, errors := getBundle(bundleCreatedResponse.Self, func(body []byte) {
				fileBytes = body
			})

			IsNil(errors)

			Expect(response.StatusCode).Should(Equal(200))

			Expect(fileBytes).Should(Equal(testPayload))

		})

		FIt("List Revisions", func() {

			bundleName := "test"

			revisionCreatedResponses := []*server.BundleCreatedResponse{}

			//create 5 unique revisions
			for i := uint32(0); i < 5; i++ {

				data := GenerateBinaryFromInt(i)

				response, bundleCreatedResponse, err := uploadBundle(testServer, bundleName, bytes.NewReader(data))

				IsNil(err)

				Expect(response.StatusCode).Should(Equal(http.StatusCreated))

				revisionCreatedResponses = append(revisionCreatedResponses, bundleCreatedResponse)
			}

			response, revisions, errors := getRevisions(testServer, bundleName, "", 2)

			IsNil(errors)

			Expect(response.StatusCode).Should(Equal(http.StatusOK))

			Expect(revisions.Cursor).ShouldNot(BeEmpty())

			Expect(len(revisions.Revisions)).Should(Equal(2))

			Expect(revisions.Revisions[0].Revision).Should(Equal(revisionCreatedResponses[0].Revision))
			Expect(revisions.Revisions[0].Self).Should(Equal(revisionCreatedResponses[0].Self))
			Expect(revisions.Revisions[0].Created).ShouldNot(BeZero())

			Expect(revisions.Revisions[1].Revision).Should(Equal(revisionCreatedResponses[1].Revision))
			Expect(revisions.Revisions[1].Self).Should(Equal(revisionCreatedResponses[1].Self))
			Expect(revisions.Revisions[1].Created).ShouldNot(BeZero())

			response, revisions, errors = getRevisions(testServer, bundleName, revisions.Cursor, 2)

			IsNil(errors)

			Expect(response.StatusCode).Should(Equal(http.StatusOK))

			Expect(revisions.Cursor).ShouldNot(BeEmpty())

			Expect(len(revisions.Revisions)).Should(Equal(2))

			Expect(revisions.Revisions[0].Revision).Should(Equal(revisionCreatedResponses[2].Revision))
			Expect(revisions.Revisions[0].Self).Should(Equal(revisionCreatedResponses[2].Self))
			Expect(revisions.Revisions[0].Created).ShouldNot(BeZero())

			Expect(revisions.Revisions[1].Revision).Should(Equal(revisionCreatedResponses[3].Revision))
			Expect(revisions.Revisions[1].Self).Should(Equal(revisionCreatedResponses[3].Self))
			Expect(revisions.Revisions[1].Created).ShouldNot(BeZero())

			response, revisions, errors = getRevisions(testServer, bundleName, revisions.Cursor, 2)

			IsNil(errors)

			Expect(response.StatusCode).Should(Equal(http.StatusOK))

			Expect(revisions.Cursor).Should(BeEmpty())

			Expect(len(revisions.Revisions)).Should(Equal(1))

			Expect(revisions.Revisions[0].Revision).Should(Equal(revisionCreatedResponses[4].Revision))
			Expect(revisions.Revisions[0].Self).Should(Equal(revisionCreatedResponses[4].Self))
			Expect(revisions.Revisions[0].Created).ShouldNot(BeZero())

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

//Upload a bundle and parse the response.  Either the bundleCreatedResponse will be returned, or the errors will
func uploadBundle(testServer *httptest.Server, bundleName string, fileData io.Reader) (*http.Response, *server.BundleCreatedResponse, *server.Errors) {

	url := testServer.URL + "/api/bundles"

	// response, body, err := newFileUploadRequest(url, bundleName, fileData)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("bundleData", "data.zip")

	IsNil(err)

	//copy over the bytes
	_, err = io.Copy(part, fileData)

	IsNil(err)

	writer.WriteField("bundleName", bundleName)

	//set the content type
	writer.FormDataContentType()

	err = writer.Close()

	IsNil(err)

	request, err := http.NewRequest("POST", url, body)

	IsNil(err)

	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	response, err := client.Do(request)

	IsNil(err)

	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		errors := &server.Errors{}
		err = json.NewDecoder(response.Body).Decode(errors)

		IsNil(err)
		return response, nil, errors

	}

	bundleCreatedResponse := &server.BundleCreatedResponse{}

	err = json.NewDecoder(response.Body).Decode(bundleCreatedResponse)

	IsNil(err)

	return response, bundleCreatedResponse, nil

}

//submit the get request.  If the response is 200, then the parser will be invoked.  Otherwise errors will be parsed
func getBundle(url string, parser func(body []byte)) (*http.Response, *server.Errors) {

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

//getRevisions get the revisions of the bundle
func getRevisions(testServer *httptest.Server, bundleName, cursor string, pageSize int) (*http.Response, *server.BundleRevisions, *server.Errors) {

	revisionsURL := fmt.Sprintf("%s/api/bundles/%s/revisions?cursor=%s&pageSize=%d", testServer.URL, bundleName, cursor, pageSize)

	request, err := http.NewRequest("GET", revisionsURL, nil)

	IsNil(err)

	client := &http.Client{}

	response, err := client.Do(request)

	IsNil(err)

	defer response.Body.Close()

	// fmt.Printf("Response body is %s", string(responseBody))

	//only parse our org if it's a successfull response code
	if response.StatusCode != 200 {

		errorResponse := &server.Errors{}

		err := json.NewDecoder(response.Body).Decode(errorResponse)
		IsNil(err)

		return response, nil, errorResponse

	}

	revisionsResponse := &server.BundleRevisions{}

	err = json.NewDecoder(response.Body).Decode(revisionsResponse)
	IsNil(err)

	return response, revisionsResponse, nil

}

//responseBodyAsBytes get the response body and close it properly
func resposneBodyAsBytes(response *http.Response) []byte {
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	IsNil(err)

	return bodyBytes

}

type noOpTestAuth struct{}

func (n *noOpTestAuth) VerifyOAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(rw, r)
	})
}
