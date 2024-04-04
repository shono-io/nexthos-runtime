FROM golang:1.22 AS build

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /go/src/app

ARG GITHUB_USERNAME
ARG GITHUB_TOKEN

# Update dependencies: On unchanged dependencies, cached layer will be reused
COPY go.* /go/src/app
RUN go mod download

# Build
COPY . /go/src/app/
RUN go build -o nexthos_runtime

# Pack
FROM debian:bookworm-slim AS package

LABEL maintainer="Shono <code@shono.io>"

WORKDIR /

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /go/src/app/nexthos_runtime .

ENTRYPOINT ["/nexthos_runtime"]

EXPOSE 8080