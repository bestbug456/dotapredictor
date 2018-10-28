# Use an official Golang runtime as build env
FROM golang:1.11-alpine AS build-env
RUN apk add git
WORKDIR /app
ADD . /app
RUN cd /app && export CGO_ENABLED=0 && go build -o dotapredictor

# use alpine for run the application
FROM alpine
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build-env /app/dotapredictor /app

ENTRYPOINT ./dotapredictor