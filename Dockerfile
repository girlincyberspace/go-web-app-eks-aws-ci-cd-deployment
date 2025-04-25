FROM golang:1.23 AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o url-shortener main.go


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/url-shortener .
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./url-shortener"]