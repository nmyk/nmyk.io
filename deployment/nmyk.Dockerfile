FROM debian:testing
LABEL maintainer="Nick Mykins <nick@nmyk.io>"

RUN apt update && apt install \
    golang \
    build-essential \
    -y

WORKDIR /app
COPY main.go ./
COPY web/ ./web/

RUN mkdir /build
RUN go build -v -o /build/nmyk

ENTRYPOINT ["/build/nmyk"]
