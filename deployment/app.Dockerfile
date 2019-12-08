FROM debian:testing
LABEL maintainer="Nick Mykins <nick@nmyk.io>"

RUN apt update && apt install \
    golang \
    build-essential \
    -y

WORKDIR /app
ADD pkg ./
COPY web/index.html ./web/

RUN mkdir /build
RUN go build -v -o /build/app

ENTRYPOINT ["/build/app"]