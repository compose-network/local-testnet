# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/op-proposer ./op-proposer/cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/op-proposer /usr/local/bin/op-proposer
ENTRYPOINT ["/usr/local/bin/op-proposer"]
