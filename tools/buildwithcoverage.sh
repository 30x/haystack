#!/bin/sh


export GOOGLE_APPLICATION_CREDENTIALS="/go/src/github.com/30x/haystack/build/application_default_credentials.json"

source /go/src/github.com/30x/haystack/build/env.sh

#Run our tests

echo "PROJECTID = $PROJECTID"
echo "GOOGLE_APPLICATION_CREDENTIALS = $GOOGLE_APPLICATION_CREDENTIALS"

coverMode="atomic"

cd /go/src/github.com/30x/haystack/

dirs="./api ./storage"

echo "mode: $coverMode" > coverage.txt


for d in $dirs; do
    go test -v -coverprofile=profile.out -covermode=$coverMode $d
    
    if [ $? -ne 0 ]; then
        echo "Tests failed"
        exit 1
    fi

    if [ -f profile.out ]; then
        tail -n+2 profile.out >> coverage.txt
        rm profile.out
    fi
done

go tool cover -html=coverage.txt -o cover.html

go build -a -installsuffix cgo -ldflags '-w' -o build/haystack .