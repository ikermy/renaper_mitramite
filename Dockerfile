FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bot ./cmd/

FROM alpine:3.19
RUN apk --no-cache add ca-certificates chromium chromium-chromedriver
WORKDIR /app
COPY --from=builder /app/bot ./bot

ENV CHROMIUM_PATH=/usr/bin/chromium-browser
EXPOSE 8080
CMD ["./bot"]
