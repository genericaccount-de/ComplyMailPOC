.PHONY: build-all test-all lint-all run-api run-proxy clean

# --- Build ---
build-all: build-api build-proxy

build-api:
	cd backend && go build -o ../bin/api ./cmd/api

build-proxy:
	cd smtp-proxy && go build -o ../bin/proxy ./cmd/proxy

# --- Test ---
test-all:
	cd backend && go test ./...
	cd smtp-proxy && go test ./...

# --- Lint ---
lint-all:
	cd backend && golangci-lint run ./...
	cd smtp-proxy && golangci-lint run ./...

# --- Run (local) ---
run-api:
	cd backend && go run ./cmd/api

run-proxy:
	cd smtp-proxy && go run ./cmd/proxy

# --- Docker ---
docker-up:
	docker compose up --build

docker-down:
	docker compose down

# --- Clean ---
clean:
	rm -rf bin/
