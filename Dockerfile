# ---------- Build stage ----------
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /app/bin/server ./cmd/server


# ---------- Runtime stage ----------
FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    wget

WORKDIR /app

COPY --from=builder \
    /app/bin/server \
    ./server

EXPOSE 8080
EXPOSE ${APP_PORT} 

CMD ["./server"]