# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.18-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /extension-prometheus

##
## Runtime
##
FROM alpine:3.16

WORKDIR /

COPY --from=build /extension-prometheus /extension-prometheus

EXPOSE 8084

ENTRYPOINT ["/extension-prometheus"]
