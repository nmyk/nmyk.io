FROM golang:alpine
LABEL maintainer="Nick Mykins <nick@nmyk.io>"

WORKDIR /app
COPY . /app/

RUN mkdir /build
RUN go build -v -o /build/app

ENTRYPOINT ["/build/app"]
