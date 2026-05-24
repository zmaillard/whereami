# syntax=docker/dockerfile:1
FROM golang:1.26 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download


COPY . ./

ENV CGO_ENABLED=1
ARG VERSION_APP=0.0.0
ENV VERSION ${VERSION_APP}

RUN mkdir -p /assets

RUN go build \
    --ldflags "-X 'main.Version=${VERSION_APP}'" \
    -o /main main.go

##
## Deploy
##
FROM alpine:3.23

RUN apk add libspatialite

WORKDIR /

COPY --from=build /main /main
COPY --from=build /assets /assets
#COPY ./whereami.db /whereami.db

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/main"]
