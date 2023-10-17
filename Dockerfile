# Build Stage
FROM golang:1.21.3-alpine3.18 AS BuildStage
RUN apk update && apk add --no-cache musl-dev gcc build-base
RUN apk add --no-cache make
# Create user
RUN adduser -D -g '' user

# Copy sources
COPY internal /fantaleghe/internal
COPY go.mod go.sum *.go Makefile /fantaleghe/
# Build
WORKDIR /fantaleghe
RUN go mod tidy && make build

# Deploy Stage
FROM alpine:latest
WORKDIR /
COPY --from=BuildStage /fantaleghe/bin /bin
RUN adduser -D -g '' user
ENV DOCKER_ENV=true

USER user:user
ENTRYPOINT ["/bin/fantalegheGO"]

EXPOSE 8080