VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  = -s -w -X github.com/nextlevelbuilder/goclaw/cmd.Version=$(VERSION)
BINARY   = goclaw

.PHONY: build run clean version up down logs test vet lint-web ci setup

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

version:
	@echo $(VERSION)

COMPOSE = docker compose -f docker-compose.yml -f docker-compose.managed.yml -f docker-compose.selfservice.yml

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f goclaw

test:
	go test -race ./...

vet:
	go vet ./...

lint-web:
	cd ui/web && pnpm install --frozen-lockfile && pnpm build

setup:
	go mod download
	cd ui/web && pnpm install --frozen-lockfile

ci: build test vet lint-web
