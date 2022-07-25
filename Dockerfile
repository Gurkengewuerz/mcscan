# Start by building the application.
FROM golang:1.17-buster as build

WORKDIR /go/src/app
ADD . /go/src/app

RUN apt-get update && apt-get install -y \
    libpcap-dev \
 && rm -rf /var/lib/apt/lists/*

RUN go get -d -v ./...

RUN go build -o /go/bin/app

# Now copy it into our base image.
FROM debian:bullseye

RUN apt-get update && apt-get install -y \
    masscan \
    libpcap0.8 \
    libpcap-dev \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /run

COPY --from=build /go/bin/app /run/app
COPY --from=build /go/src/app/exclude.conf /run/exclude.conf

CMD ["/run/app"]