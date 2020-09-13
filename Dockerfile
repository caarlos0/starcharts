# GitHub:       https://github.com/lucmichalski
FROM golang:1.15-alpine AS build

ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

ARG CGO=1
ENV CGO_ENABLED=${CGO}
ENV GOOS=linux
ENV GO111MODULE=on

WORKDIR /go/src/github.com/caarlos0/starcharts

COPY . /go/src/github.com/caarlos0/starcharts

RUN apk update && \
    apk add --no-cache git ca-certificates make

RUN go build -ldflags "-extldflags=-static -extldflags=-lm" -o /go/bin/starcharts

FROM paper2code/starcharts:latest as twint

FROM alpine:3.12

COPY --from=build /go/bin/starcharts /usr/bin/starcharts

RUN apk update && \
    apk add --no-cache ca-certificates nano bash

VOLUME /data
WORKDIR /data

# Expose port for live server
EXPOSE 3000

CMD ["/usr/bin/starcharts"]

