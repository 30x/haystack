.PHONY: test

VERSION=0.0.1-dev


make-push: test compile-linux build-image push-to-hub

test:
	go test -v ./api && go test -v ./storage 

view-coverage:
	go tool cover -html=coverage.out

compile-linux:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o build/haystack .

build-image:
#Copy over the required credentials file
	sh tools/createdockerfiles.sh
	#Run the docker build
	docker build -t thirtyx/haystack .

push-to-hub:
	docker tag thirtyx/haystack thirtyx/haystack:$(VERSION)
	docker push thirtyx/haystack:$(VERSION)
