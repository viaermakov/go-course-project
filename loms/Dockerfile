FROM golang:1.22-alpine as builder

WORKDIR /app

COPY go.mod go.mod

RUN go mod download

COPY . .

RUN go build ./cmd/server

FROM alpine

COPY --from=builder /app/bin/server /server
COPY --from=builder /app/.env /.env

CMD [ "/server" ]
