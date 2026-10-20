FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /payment-mock .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /payment-mock /payment-mock

EXPOSE 8080
ENTRYPOINT ["/payment-mock"]
