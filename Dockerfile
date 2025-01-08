FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o audiobook-organizer

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/audiobook-organizer .

ENTRYPOINT ["/app/audiobook-organizer"]