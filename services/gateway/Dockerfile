# Étape 1 : build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Dépendances
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy

# Code source
COPY . .

# Download all dependencies including swaggo
RUN go mod download
RUN go mod tidy

# Generate Swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/api/main.go --output docs

# Build - compile depuis cmd/api/main.go (nouvelle structure)
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/api/main.go

# Étape 2 : runtime
FROM alpine:latest

WORKDIR /app

# Certificats SSL (important)
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/app .
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./app"]
