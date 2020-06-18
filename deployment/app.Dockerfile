FROM golang:alpine
LABEL maintainer="Nick Mykins <nick@nmyk.io>"

COPY . /app/
RUN rmdir /app/deployment

WORKDIR /app/pkg

ENV GOPATH=/go:/app/pkg
ENV GOBIN=$GOPATH/bin

RUN apk add --no-cache git
RUN go get -v .

RUN mkdir /build
RUN go build -v -o /build/app

WORKDIR /app
ENTRYPOINT ["/build/app"]
