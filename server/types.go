package server

//BundleCreatedResponse the created response for the api
type BundleCreatedResponse struct {
	Revision string `json:"revision"`
	Self     string `json:"self"`
}

//Errors to return
type Errors []string
