FROM golang:1.21.0 AS builder

WORKDIR /app

COPY . .
RUN go mod download

RUN OS="linux" make build
