package httputil

//Errors to return
type Errors []string

//HasErrors return true if there are errors
func (e Errors) HasErrors() bool {
	return len(e) > 0
}
