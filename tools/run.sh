#!/bin/sh
source $GOPATH/src/github.com/30x/haystack/build/env.sh

#Run our tests
echo "PROJECTID = $PROJECTID"
echo "GOOGLE_APPLICATION_CREDENTIALS = $GOOGLE_APPLICATION_CREDENTIALS"

cd $GOPATH/src/github.com/30x/haystack/
go build -a -installsuffix cgo -ldflags '-w' -o build/haystack .
build/haystack