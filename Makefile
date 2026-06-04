.PHONY: dev dev-backend dev-frontend build docker-build helm-install

# Run backend with mock data (no cluster needed)
dev-backend:
	cd backend && USE_MOCK=true go run .

# Run frontend dev server (proxies /api to localhost:8080)
dev-frontend:
	cd frontend && npm install && npm run dev

# Build frontend and copy dist into backend embed path, then run backend
dev:
	cd frontend && npm install && npm run build && cp -r dist ../backend/frontend/dist
	cd backend && USE_MOCK=true go run .

# Build the Docker image
docker-build:
	docker build -t kubera:latest .

# Run go mod tidy
tidy:
	cd backend && go mod tidy

# Install via Helm (set GEMINI_API_KEY env var first)
helm-install:
	helm install kubera ./helm/kubera \
		--namespace kubera \
		--create-namespace \
		--set gemini.apiKey="$(GEMINI_API_KEY)"
