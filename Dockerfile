# Build stage
FROM golang:1.25-alpine AS build
WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY backend ./backend

RUN CGO_ENABLED=0 GOOS=linux go build -o /vecai-api ./backend/cmd/api

# Runtime stage
FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 10001 appuser

COPY --from=build /vecai-api /app/vecai-api
COPY backend/db/migrations /app/backend/db/migrations

RUN mkdir -p /app/backend/uploads && chown -R appuser:appuser /app
USER appuser

ENV PORT=8080
ENV MIGRATIONS_PATH=/app/backend/db/migrations
ENV UPLOAD_DIR=/app/backend/uploads

EXPOSE 8080
CMD ["/app/vecai-api"]
