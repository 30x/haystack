# FROM scratch
FROM golang:1.7.4-alpine

COPY . /go/src/github.com/30x/haystack

RUN /go/src/github.com/30x/haystack/tools/buildwithcoverage.sh && cp /go/src/github.com/30x/haystack/build/haystack /

EXPOSE 5280

ENTRYPOINT [ "/haystack" ]