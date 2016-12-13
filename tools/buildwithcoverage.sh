#!/bin/sh


source buildtime/env.sh
#Run our tests

coverMode="atomic"

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