FROM golang:alpine
LABEL maintainer="Nick Mykins <nick@nmyk.io>"

COPY . /app/
RUN rmdir /app/deployment

WORKDIR /app/pkg
RUN mkdir /build
RUN go build -v -o /build/app

WORKDIR /app
ENTRYPOINT ["/build/app"]
