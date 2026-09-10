# syntax=docker/dockerfile:1

ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS app-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/adex ./cmd/adex

FROM golang:${GO_VERSION}-alpine AS goose-build
ARG GOOSE_VERSION=v3.24.1
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

FROM alpine:3.22 AS app
RUN addgroup -S adex && adduser -S -G adex adex
COPY --from=app-build /out/adex /usr/local/bin/adex
USER adex
EXPOSE 8080
ENTRYPOINT ["adex"]

FROM alpine:3.22 AS migrator
COPY --from=goose-build /go/bin/goose /usr/local/bin/goose
COPY migrations /migrations
ENTRYPOINT ["goose", "-dir", "/migrations"]
