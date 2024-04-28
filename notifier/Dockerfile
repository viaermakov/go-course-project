FROM golang:1.22-alpine AS build-stage

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/server

FROM alpine:3 AS build-release-stage

WORKDIR /

COPY --from=build-stage /app/server /server
COPY --from=build-stage /app/.env /.env

ENTRYPOINT ["/server"]
