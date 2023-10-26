# syntax=docker/dockerfile:1

FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

## Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY *.go ./

RUN go build -o /web_server

EXPOSE 8080

CMD [ "/web_server" ]