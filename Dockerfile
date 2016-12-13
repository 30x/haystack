# FROM scratch
FROM golang:1.7.4-alpine



#Deliberately done as it's own command so this is cached and only happens on first build
RUN apk add --update build-base --virtual haystack-build

COPY . /go/src/github.com/30x/haystack

#Copy over the creds for running google tests
COPY .config/gcloud/application_default_credentials.json /users/root/.config/gcloud/application_default_credentials.json

RUN source /go/src/github.com/30x/haystack/environment.sh
RUN echo $PROJECTID
RUN echo "User is"
RUN whoami

RUN \
    (cd /go/src/github.com/30x/haystack; make; cp build/haystack /) \
 && rm -r /go \ 
 && apk del haystack-build

EXPOSE 5280

ENTRYPOINT [ "/haystack" ]