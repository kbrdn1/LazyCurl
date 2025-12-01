.PHONY: build run clean test install dev help

# Variables
BINARY_NAME=lazycurl
BINARY_PATH=bin/$(BINARY_NAME)
CMD_PATH=./cmd/lazycurl
GO=go
GOFLAGS=-v

# Default target
all: build

## help: Affiche cette aide
help:
	@echo "Commandes disponibles:"
	@echo "  make build     - Compile l'application"
	@echo "  make run       - Compile et lance l'application"
	@echo "  make clean     - Nettoie les binaires et caches"
	@echo "  make test      - Lance les tests"
	@echo "  make install   - Installe l'application globalement"
	@echo "  make dev       - Mode développement avec live reload"
	@echo "  make deps      - Télécharge les dépendances"

## build: Compile l'application
build:
	@echo "🔨 Compilation de $(BINARY_NAME)..."
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_PATH) $(CMD_PATH)
	@echo "✅ Binaire créé: $(BINARY_PATH)"

## run: Compile et lance l'application
run: build
	@echo "🚀 Lancement de $(BINARY_NAME)..."
	@$(BINARY_PATH)

## clean: Nettoie les binaires et caches
clean:
	@echo "🧹 Nettoyage..."
	@rm -rf bin/
	@rm -rf dist/
	@$(GO) clean
	@echo "✅ Nettoyage terminé"

## test: Lance les tests
test:
	@echo "🧪 Lancement des tests..."
	@$(GO) test -v ./...

## test-coverage: Lance les tests avec coverage
test-coverage:
	@echo "🧪 Lancement des tests avec coverage..."
	@$(GO) test -coverprofile=coverage.txt -covermode=atomic ./...
	@$(GO) tool cover -html=coverage.txt -o coverage.html
	@echo "✅ Coverage report: coverage.html"

## install: Installe l'application globalement
install:
	@echo "📦 Installation de $(BINARY_NAME)..."
	@$(GO) install $(CMD_PATH)
	@echo "✅ $(BINARY_NAME) installé dans $(GOPATH)/bin"

## deps: Télécharge les dépendances
deps:
	@echo "📥 Téléchargement des dépendances..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "✅ Dépendances installées"

## dev: Mode développement avec live reload (nécessite air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "❌ 'air' n'est pas installé. Installation..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

## fmt: Formate le code
fmt:
	@echo "✨ Formatage du code..."
	@$(GO) fmt ./...
	@echo "✅ Code formaté"

## lint: Vérifie le code avec golangci-lint
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "🔍 Analyse du code..."; \
		golangci-lint run; \
	else \
		echo "❌ golangci-lint n'est pas installé"; \
		echo "Installation: https://golangci-lint.run/usage/install/"; \
	fi

## build-all: Compile pour toutes les plateformes
build-all:
	@echo "🔨 Compilation multi-plateforme..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GO) build -o dist/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	GOOS=linux GOARCH=arm64 $(GO) build -o dist/$(BINARY_NAME)-linux-arm64 $(CMD_PATH)
	GOOS=darwin GOARCH=amd64 $(GO) build -o dist/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build -o dist/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build -o dist/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "✅ Binaires créés dans dist/"

## version: Affiche la version de Go
version:
	@$(GO) version
