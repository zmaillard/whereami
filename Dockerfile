# syntax=docker/dockerfile:1
FROM golang:1.26-bookworm AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

ENV CGO_ENABLED=1
ARG VERSION_APP=0.0.0
ENV VERSION ${VERSION_APP}


RUN go build \
    --ldflags "-X 'main.Version=${VERSION_APP}'" \
    -o /main main.go


##
## Deploy
##
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y apt-file libsqlite3-mod-spatialite && rm -rf /var/lib/apt/lists/*
WORKDIR /

COPY --from=build /main /main
COPY --from=build /app/assets /assets

EXPOSE 8080

ENTRYPOINT ["/main"]
