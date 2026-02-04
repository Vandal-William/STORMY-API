# Étape 1 : build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Dépendances
COPY go.mod go.sum ./
RUN go mod download

# Code source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o app

# Étape 2 : runtime
FROM alpine:latest

WORKDIR /app

# Certificats SSL (important)
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/app .

EXPOSE 3002

CMD ["./app"]
