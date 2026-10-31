.PHONY: build run dev frontend-dev frontend/dist frontend-bundle-check deploy-frontend install-air clean container container-run

# Which env file to export. Override for another local profile:
#   make run ENV_FILE=.env.staging
ENV_FILE ?= .env.development

build:
	CGO_ENABLED=0 go build -o server ./cmd/server

# Vite loads `.env.<mode>`, and `vite build` runs in production mode — so with
# the file named .env.development a local `npm run build` found no
# VITE_CLERK_PUBLISHABLE_KEY and the bundle rendered the "key is not set"
# message. Exporting the env file into the build process is what actually
# works, and it is mode-independent.
#
# Declared .PHONY on purpose: as a real `frontend/dist` directory target it had
# no prerequisites, so after the first build make always reported it
# "up-to-date" and skipped the build — which is how a stale, keyless bundle got
# deployed. It now rebuilds every time, then asserts the key is actually
# inlined.
frontend/dist:
	cd frontend && npm ci && export $$(grep -v '^#' $(CURDIR)/$(ENV_FILE) | xargs) && npm run build
	./scripts/frontend-bundle-check.sh frontend/dist

# Manual Cloudflare SPA deploy. Build (guarded) first, then ship — never deploy
# a dist that skipped the key check.
deploy-frontend: frontend/dist
	npx wrangler deploy

run: build
	export $$(grep -v '^#' $(ENV_FILE) | xargs) && ./server

dev:
	PATH="$$(go env GOPATH)/bin:$$PATH" air

# The Go binary serves the API only; the SPA is a separate process (Vite in dev,
# Cloudflare Pages in prod). Run both during local development.
frontend-dev:
	cd frontend && npm run dev

install-air:
	go install github.com/air-verse/air@latest

clean:
	rm -f server
	rm -rf frontend/node_modules frontend/dist

container:
	@echo "Usage: make container IMAGE=paperviz TAG=latest"
	@echo "       make container IMAGE=paperviz TAG=v1"
	podman build -t $(if $(IMAGE),$(IMAGE),paperviz):$(if $(TAG),$(TAG),latest) -f build/Containerfile .

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
