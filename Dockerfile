FROM golang:1.15-alpine AS build
MAINTAINER paper2code <contact@paper2code.com>

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

FROM alpine:3.12
MAINTAINER paper2code <contact@paper2code.com>

COPY --from=build /go/bin/starcharts /usr/bin/starcharts

RUN apk update && \
    apk add --no-cache ca-certificates nano bash

WORKDIR /opt/starcharts
VOLUME /opt/starcharts/data

# Expose port for live server
EXPOSE 3000

CMD ["starcharts"]

