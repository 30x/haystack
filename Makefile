VERSION=0.0.1-dev

.PHONY: test


make-push: test compile-linux build-image push-to-hub

test:
    go test $((glide novendor))-coverprofile=coverage.out 

view-coverage:
    go tool cover -html=coverage.out

compile-linux:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o build/haystack .

build-image:
	docker build -t thirtyx/haystack .

push-to-hub:
	docker tag thirtyx/haystack thirtyx/haystack:$(VERSION)
	docker push thirtyx/haystack:$(VERSION)