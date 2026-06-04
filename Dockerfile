# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# Copy compiled frontend into embed path expected by main.go
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o kubera .

# Stage 3: Final minimal image
FROM gcr.io/distroless/static:nonroot
COPY --from=backend-builder /app/kubera /kubera
EXPOSE 8080
ENTRYPOINT ["/kubera"]
