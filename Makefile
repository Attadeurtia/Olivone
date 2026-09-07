# Raccourcis de développement Olivone.
# `make help` liste les cibles.

.PHONY: help dev down logs web server test tidy build

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

dev: ## Lance la stack locale (app + Mailpit + GreenMail) via Docker
	docker compose -f docker-compose.dev.yml up --build

down: ## Arrête la stack locale
	docker compose -f docker-compose.dev.yml down

logs: ## Suit les logs de l'app
	docker compose -f docker-compose.dev.yml logs -f app

web: ## Construit le frontend (web/dist)
	cd web && npm install && npm run build

server: ## Lance le serveur Go en natif (placeholder si web non construit)
	cd server && OLIVONE_ENV=dev OLIVONE_DATA_DIR=../data go run ./cmd/olivone

test: ## Lance les tests Go
	cd server && go test ./...

tidy: ## Met à jour les dépendances Go
	cd server && go mod tidy

build: ## Build complet (frontend + binaire Go dans server/bin/olivone)
	cd web && npm install && npm run build
	cd server && CGO_ENABLED=0 go build -o bin/olivone ./cmd/olivone
