# Use an official Golang runtime as build env
FROM golang:1.10-alpine AS build-env
WORKDIR /app
ADD . /app
RUN cd /app && go get -t -v ./... && go build -o dotapredictor

# use alpine for run the application
FROM alpine
RUN apk update && apk add --no-cache ca-certificates apache2-utils && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build-env /app/dotapredictor /app

EXPOSE 8080
ENTRYPOINT ./dotapredictor