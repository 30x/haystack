#The google project ID to use to access the cloud storage api
export PROJECTID=""

#The google cloud storage bucket name
export BUCKET_NAME=""

#The port to run on
export PORT="5280"

#The URL to the SSO symetric key. In production it's
export SSO_KEY_URL=""

export GOOGLE_APPLICATION_CREDENTIALS="$GOPATH/src/github.com/30x/haystack/build/svc.json"