.PHONY: build run dev install-air clean container container-run

build: frontend/dist
	CGO_ENABLED=0 go build -o server ./cmd/server

frontend/dist:
	cd frontend && npm ci && npm run build

run: build
	export $$(grep -v '^#' .env | xargs) && ./server

dev:
	PATH="$$(go env GOPATH)/bin:$$PATH" air

install-air:
	go install github.com/air-verse/air@latest

clean:
	rm -f server
	rm -rf frontend/node_modules frontend/dist

container:
	@echo "Usage: make container IMAGE=paperviz TAG=latest"
	@echo "       make container IMAGE=paperviz TAG=v1"
	podman build -t $(if $(IMAGE),$(IMAGE),paperviz):$(if $(TAG),$(TAG),latest) -f Containerfile .

container-run:
	@test -n "$(PORT)" || PORT=8080; \
	test -n "$(DB_PATH)" || DB_PATH=/data/paperviz.db; \
	podman run -d \
		--name paperviz \
		-p $$PORT:8080 \
		-v paperviz-data:/data \
		-e GEMINI_API_KEY="$(GEMINI_API_KEY)" \
		-e GEMINI_MODEL="$(GEMINI_MODEL)" \
		-e DATABASE_PATH="$$DB_PATH" \
		$(if $(IMAGE),$(IMAGE),paperviz):$(if $(TAG),$(TAG),latest)
	@echo "PaperViz running on http://localhost:$$PORT"

container-stop:
	podman stop paperviz && podman rm paperviz
