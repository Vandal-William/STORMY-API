# Étape 1 : build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Dépendances
COPY go.mod go.sum* ./
RUN go mod download

# Code source
COPY . .

# Installer gocql si manquant
RUN go get github.com/gocql/gocql@v1.6.0 || true

# Nettoyer les dépendances
RUN go mod tidy

# Build - compile depuis cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/api

# Étape 2 : runtime
FROM alpine:latest

WORKDIR /app

# Certificats SSL (important)
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/app .

EXPOSE 3001

CMD ["./app"]
