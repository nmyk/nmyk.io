FROM golang:alpine
LABEL maintainer="Nick Mykins <nick@nmyk.io>"

WORKDIR /app
ADD pkg ./
COPY web/index.html ./web/

RUN mkdir /build
RUN go build -v -o /build/app

ENTRYPOINT ["/build/app"]
