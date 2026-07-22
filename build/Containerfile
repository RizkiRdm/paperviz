# Stage 1 — Build frontend
FROM node:24-alpine AS frontend
WORKDIR /build
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2 — Build Go server
FROM docker.io/golang:1.24-alpine AS backend
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /build/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

# Stage 3 — Runtime (3.3 MB + binary = ~20 MB)
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -h /app paperviz
COPY --from=backend /server /app/server
COPY --from=frontend /build/dist /app/frontend/dist
WORKDIR /app
USER paperviz
EXPOSE 8080
CMD ["/app/server"]
