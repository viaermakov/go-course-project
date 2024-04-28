FROM golang:1.22-alpine as builder

WORKDIR /app

COPY go.mod go.mod

RUN go mod download

COPY . .

RUN go build ./cmd/cart-server

FROM alpine

COPY --from=builder /app/bin/cart-server /cart-server
COPY --from=builder /app/.env /.env

CMD [ "/cart-server" ]
