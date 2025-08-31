FROM golang:1.24.5-alpine AS builder

RUN apk add --no-cache git

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=direct \
    GOSUMDB=off

WORKDIR /app

COPY . .
RUN go mod download

RUN go build -o user-service ./cmd/main.go

FROM alpine:3.22

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/user-service .

EXPOSE 8080

ENTRYPOINT ["./user-service"]