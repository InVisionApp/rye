FROM golang:1.6-alpine

RUN apk update && apk add make git

RUN mkdir -p /go/src/github.com/InVisionApp/rye
ADD . /go/src/github.com/InVisionApp/rye
WORKDIR /go/src/github.com/InVisionApp/rye

RUN go get github.com/cactus/go-statsd-client/statsd && \
    go get github.com/Sirupsen/logrus && \
    go get github.com/gorilla/mux && \
    go get github.com/onsi/ginkgo && \
    go get github.com/onsi/gomega 

ENV CGO_ENABLED=0
ENV GOOS=linux
