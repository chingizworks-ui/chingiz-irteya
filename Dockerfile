FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /bin/stockpilot ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/stockpilot /usr/local/bin/stockpilot
COPY migrations ./migrations
EXPOSE 8080
CMD ["stockpilot"]
