package api_test

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

	"github.com/30x/haystack/api"
	"github.com/30x/haystack/storage"
	. "github.com/30x/haystack/test"
	uuid "github.com/satori/go.uuid"

	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("server", func() {

	var testServer *httptest.Server

	TestApi := func() {
		It("Bundle Upload and Get", func() {
			testPayload := CreateFakeBinary(10)

			bundleName := "test" + uuid.NewV1().String()

			response, bundleCreatedResponse, err := uploadBundle(testServer, bundleName, bytes.NewReader(testPayload))

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(http.StatusCreated))

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

		It("List Revisions", func() {

			bundleName := "test" + uuid.NewV1().String()

			revisionCreatedResponses := []*api.BundleCreatedResponse{}

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

		It("Test Tagging", func() {

			bundleName := "test" + uuid.NewV1().String()

			response, bundleCreatedResponse1, err := uploadBundle(testServer, bundleName, bytes.NewReader(CreateFakeBinary(10)))

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(http.StatusCreated))

			response, bundleCreatedResponse2, err := uploadBundle(testServer, bundleName, bytes.NewReader(CreateFakeBinary(9)))

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(http.StatusCreated))

			//now create 2 tags for each revision and get them back

			tag := "test1"

			response, tagResponse, err := tagBundle(testServer, bundleName, bundleCreatedResponse1.Revision, tag)

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(http.StatusCreated))

			Expect(tagResponse.Tag).Should(Equal(tag))

			Expect(tagResponse.Revision).Should(Equal(bundleCreatedResponse1.Revision))

			expectedUrl := fmt.Sprintf("%s/api/bundles/%s/tags/%s", testServer.URL, bundleName, tag)

			Expect(tagResponse.Self).Should(Equal(expectedUrl))

			//now do a get on the URL, and make sure it's expected

			response, tagInfo, err := getTagInfo(tagResponse.Self)

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(200))

			Expect(tagInfo.Tag).Should(Equal(tag))
			Expect(tagInfo.Revision).Should(Equal(bundleCreatedResponse1.Revision))
			Expect(tagInfo.Self).Should(Equal(expectedUrl))

			//Tag the second rev
			tag = "test2"

			response, tagResponse, err = tagBundle(testServer, bundleName, bundleCreatedResponse2.Revision, tag)

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(http.StatusCreated))

			Expect(tagResponse.Tag).Should(Equal(tag))

			Expect(tagResponse.Revision).Should(Equal(bundleCreatedResponse2.Revision))

			expectedUrl = fmt.Sprintf("%s/api/bundles/%s/tags/%s", testServer.URL, bundleName, tag)

			Expect(tagResponse.Self).Should(Equal(expectedUrl))

			//do a get
			response, tagInfo, err = getTagInfo(tagResponse.Self)

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(200))

			Expect(tagInfo.Tag).Should(Equal(tag))
			Expect(tagInfo.Revision).Should(Equal(bundleCreatedResponse2.Revision))
			Expect(tagInfo.Self).Should(Equal(expectedUrl))

		})

		It("Test Tag Paging", func() {

			bundleName := "test" + uuid.NewV1().String()

			response, bundleCreatedResponse, err := uploadBundle(testServer, bundleName, bytes.NewReader(CreateFakeBinary(10)))

			IsNil(err)

			//now create 2 tags for each revision and get them back

			tagResponses := []*api.TagInfo{}

			//create 5 tags
			for i := 0; i < 5; i++ {

				response, tagResponse, err := tagBundle(testServer, bundleName, bundleCreatedResponse.Revision, strconv.Itoa(i))

				IsNil(err)

				Expect(response.StatusCode).Should(Equal(http.StatusCreated))

				tagResponses = append(tagResponses, tagResponse)
			}

			response, tags, err := getTags(testServer, bundleName, "", 2)

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(200))

			//check the cursor is not empty
			Expect(tags.Cursor).ShouldNot(BeEmpty())
			Expect(len(tags.Tags)).Should(Equal(2))

			//Assert our values are what we'd expect
			Expect(tags.Tags[0].Tag).Should(Equal(tagResponses[0].Tag))
			Expect(tags.Tags[0].Revision).Should(Equal(tagResponses[0].Revision))
			Expect(tags.Tags[0].Self).Should(Equal(tagResponses[0].Self))

			//Tags
			Expect(tags.Tags[1].Tag).Should(Equal(tagResponses[1].Tag))
			Expect(tags.Tags[1].Revision).Should(Equal(tagResponses[1].Revision))
			Expect(tags.Tags[1].Self).Should(Equal(tagResponses[1].Self))

			response, tags, err = getTags(testServer, bundleName, tags.Cursor, 2)

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(200))

			//check the cursor is not empty
			Expect(tags.Cursor).ShouldNot(BeEmpty())
			Expect(len(tags.Tags)).Should(Equal(2))

			//Assert our values are what we'd expect
			Expect(tags.Tags[0].Tag).Should(Equal(tagResponses[2].Tag))
			Expect(tags.Tags[0].Revision).Should(Equal(tagResponses[2].Revision))
			Expect(tags.Tags[0].Self).Should(Equal(tagResponses[2].Self))

			//Tags
			Expect(tags.Tags[1].Tag).Should(Equal(tagResponses[3].Tag))
			Expect(tags.Tags[1].Revision).Should(Equal(tagResponses[3].Revision))
			Expect(tags.Tags[1].Self).Should(Equal(tagResponses[3].Self))

			response, tags, err = getTags(testServer, bundleName, tags.Cursor, 2)

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(200))

			//check the cursor is not empty
			Expect(tags.Cursor).Should(BeEmpty())
			Expect(len(tags.Tags)).Should(Equal(1))

			//Assert our values are what we'd expect
			Expect(tags.Tags[0].Tag).Should(Equal(tagResponses[4].Tag))
			Expect(tags.Tags[0].Revision).Should(Equal(tagResponses[4].Revision))
			Expect(tags.Tags[0].Self).Should(Equal(tagResponses[4].Self))

		})

		It("Test Tag Delete", func() {
			bundleName := "test" + uuid.NewV1().String()

			response, bundleCreatedResponse, err := uploadBundle(testServer, bundleName, bytes.NewReader(CreateFakeBinary(10)))

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(http.StatusCreated))

			tag := "test1"

			response, tagResponse, err := tagBundle(testServer, bundleName, bundleCreatedResponse.Revision, tag)

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(http.StatusCreated))

			//now do a get on the URL, and make sure it's expected

			response, tagInfo, err := getTagInfo(tagResponse.Self)

			IsNil(err)

			Expect(response.StatusCode).Should(Equal(http.StatusOK))

			Expect(tagInfo.Tag).Should(Equal(tag))
			Expect(tagInfo.Revision).Should(Equal(bundleCreatedResponse.Revision))

			//now delete it
			response, deleteResponse, err := deleteTag(tagResponse.Self)

			IsNil(err)
			Expect(response.StatusCode).Should(Equal(http.StatusOK))

			Expect(deleteResponse.Tag).Should(Equal(tag))
			Expect(deleteResponse.Revision).Should(Equal(bundleCreatedResponse.Revision))
			Expect(deleteResponse.Self).Should(Equal(tagResponse.Self))

		})
	}

	//Set up and execute the gcloud implementation for the tests.   Other implementations will define a new context with it's own setup, and execute the tests
	Context("GCloud storage", func() {

		var bucketName string
		var storageImpl storage.Storage

		BeforeSuite(func() {

			bucketName, storageImpl = CreateGCloudImpl()
			fakeOauth := &noOpTestAuth{}
			r := api.CreateRoutes(storageImpl, fakeOauth)

			testServer = httptest.NewServer(r)

		})

		AfterSuite(func() {
			RemoveGCloudTestBucket(bucketName, storageImpl)
		})

		TestApi()
	})

})

func tagBundle(testServer *httptest.Server, bundleName, revision, tag string) (*http.Response, *api.TagInfo, *api.Errors) {
	tagPayload := api.TagCreate{
		Revision: revision,
		Tag:      tag,
	}

	payload, err := json.Marshal(tagPayload)
	IsNil(err)

	url := fmt.Sprintf("%s/api/bundles/%s/tags", testServer.URL, bundleName)

	request, err := http.NewRequest("POST", url, bytes.NewReader(payload))

	IsNil(err)

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	response, err := client.Do(request)

	IsNil(err)

	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		errors := &api.Errors{}
		err = json.NewDecoder(response.Body).Decode(errors)

		IsNil(err)
		return response, nil, errors

	}

	tagCreatedResponse := &api.TagInfo{}

	err = json.NewDecoder(response.Body).Decode(tagCreatedResponse)

	IsNil(err)

	return response, tagCreatedResponse, nil

}

//getTagInfo get the tag info for the specified tag
func getTagInfo(tagUrl string) (*http.Response, *api.TagInfo, *api.Errors) {

	return performTagOp("GET", tagUrl)

}

//getTagInfo get the tag info for the specified tag
func deleteTag(tagUrl string) (*http.Response, *api.TagInfo, *api.Errors) {
	return performTagOp("DELETE", tagUrl)
}

func performTagOp(httpMethod, tagURL string) (*http.Response, *api.TagInfo, *api.Errors) {
	request, err := http.NewRequest(httpMethod, tagURL, nil)

	IsNil(err)

	request.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	response, err := client.Do(request)

	IsNil(err)

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		errors := &api.Errors{}
		err = json.NewDecoder(response.Body).Decode(errors)

		IsNil(err)
		return response, nil, errors

	}

	tagCreatedResponse := &api.TagInfo{}

	err = json.NewDecoder(response.Body).Decode(tagCreatedResponse)

	IsNil(err)

	return response, tagCreatedResponse, nil
}

//getTags get the tags of the bundle
func getTags(testServer *httptest.Server, bundleName, cursor string, pageSize int) (*http.Response, *api.TagsResponse, *api.Errors) {

	tagsUrl := fmt.Sprintf("%s/api/bundles/%s/tags?cursor=%s&pageSize=%d", testServer.URL, bundleName, cursor, pageSize)

	request, err := http.NewRequest("GET", tagsUrl, nil)

	IsNil(err)

	client := &http.Client{}

	response, err := client.Do(request)

	IsNil(err)

	defer response.Body.Close()

	// fmt.Printf("Response body is %s", string(responseBody))

	//only parse our org if it's a successfull response code
	if response.StatusCode != 200 {

		errorResponse := &api.Errors{}

		err := json.NewDecoder(response.Body).Decode(errorResponse)
		IsNil(err)

		return response, nil, errorResponse

	}

	tagsResponse := &api.TagsResponse{}

	err = json.NewDecoder(response.Body).Decode(tagsResponse)
	IsNil(err)

	return response, tagsResponse, nil

}

//Upload a bundle and parse the response.  Either the bundleCreatedResponse will be returned, or the errors will
func uploadBundle(testServer *httptest.Server, bundleName string, fileData io.Reader) (*http.Response, *api.BundleCreatedResponse, *api.Errors) {

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
		errors := &api.Errors{}
		err = json.NewDecoder(response.Body).Decode(errors)

		IsNil(err)
		return response, nil, errors

	}

	bundleCreatedResponse := &api.BundleCreatedResponse{}

	err = json.NewDecoder(response.Body).Decode(bundleCreatedResponse)

	IsNil(err)

	return response, bundleCreatedResponse, nil

}

//submit the get request.  If the response is 200, then the parser will be invoked.  Otherwise errors will be parsed
func getBundle(url string, parser func(body []byte)) (*http.Response, *api.Errors) {

	request, err := http.NewRequest("GET", url, nil)

	IsNil(err)

	client := &http.Client{}

	response, err := client.Do(request)

	IsNil(err)

	defer response.Body.Close()

	responseBody := resposneBodyAsBytes(response)

	// fmt.Printf("Response body is %s", string(responseBody))

	var errors *api.Errors

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
func getRevisions(testServer *httptest.Server, bundleName, cursor string, pageSize int) (*http.Response, *api.BundleRevisions, *api.Errors) {

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

		errorResponse := &api.Errors{}

		err := json.NewDecoder(response.Body).Decode(errorResponse)
		IsNil(err)

		return response, nil, errorResponse

	}

	revisionsResponse := &api.BundleRevisions{}

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
