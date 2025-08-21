.PHONY: help build auth-basic auth-improved precache-worker seeder load-test infra-up infra-down

help:
	@echo "Available commands:"
	@echo "  build          - Build all binaries at project root"
	@echo "  auth-basic      - Run auth-basic service on port 8080"
	@echo "  auth-improved   - Run auth-improved service on port 8081"
	@echo "  precache-worker - Run precache worker"
	@echo "  seeder         - Run user seeder (usage: make seeder N=1000)"
	@echo "  seeder-single  - Insert single user (usage: make seeder-single USERNAME=user@katakode.com PASSWORD=123)"
	@echo "  load-test      - Run k6 load tests"
	@echo "  infra-up       - Start Redis and MySQL"
	@echo "  infra-down     - Stop Redis and MySQL"
	@echo "  cache-reset    - Clear all Redis cache data"
	@echo "  memory-stats   - Show Redis memory usage statistics"

build:
	@echo "Building binaries..."
	GO111MODULE=on go build -o auth-basic-bin ./auth-basic/cmd/main.go
	GO111MODULE=on go build -o auth-improved-bin ./auth-improved/cmd/main.go
	GO111MODULE=on go build -o precache-worker-bin ./precache-worker/cmd/main.go
	GO111MODULE=on go build -o seeder-bin ./seeder/cmd/main.go
	@echo "Build complete: ./auth-basic-bin ./auth-improved-bin ./precache-worker-bin ./seeder-bin"

auth-basic: build
	@if [ ! -f .env ]; then echo "Creating .env from env.example..."; cp env.example .env; fi
	./auth-basic-bin

auth-improved: build
	@if [ ! -f .env ]; then echo "Creating .env from env.example..."; cp env.example .env; fi
	./auth-improved-bin

precache-worker: build
	@if [ ! -f .env ]; then echo "Creating .env from env.example..."; cp env.example .env; fi
	./precache-worker-bin

seeder: build
	@if [ ! -f .env ]; then echo "Creating .env from env.example..."; cp env.example .env; fi
	@if [ -z "$(N)" ]; then echo "Usage: make seeder N=1000"; exit 1; fi
	./seeder-bin -n $(N)

seeder-single: build
	@if [ ! -f .env ]; then echo "Creating .env from env.example..."; cp env.example .env; fi
	@if [ -z "$(USERNAME)" ] || [ -z "$(PASSWORD)" ]; then echo "Usage: make seeder-single USERNAME=user@katakode.com PASSWORD=123"; exit 1; fi
	./seeder-bin -username $(USERNAME) -password $(PASSWORD)

load-test:
	@echo "Running load tests..."
	@echo "Auth Basic test:"
	k6 run load-test/auth-basic.js
	# @echo "Auth Improved test:"
	# k6 run load-test/auth-improved.js

infra-up:
	docker-compose up -d

infra-down:
	docker-compose down

cache-reset:
	@echo "Resetting Redis cache..."
	docker exec substack-cache-auth-redis-1 redis-cli FLUSHALL
	@echo "Cache reset complete!"

memory-stats:
	@echo "Redis Memory Usage Statistics (auth:* keys)"
	@echo "=========================================="
	@docker exec substack-cache-auth-redis-1 redis-cli --raw EVAL "\
		local keys = redis.call('KEYS', 'auth:*'); \
		local count = #keys; \
		local total = 0; \
		local sample = 0; \
		if count > 0 then \
			sample = redis.call('MEMORY', 'USAGE', keys[1]); \
			for i=1,count do \
				total = total + redis.call('MEMORY', 'USAGE', keys[i]); \
			end; \
		end; \
		return count .. ' ' .. sample .. ' ' .. total;" 0 2>/dev/null | \
	awk '{ \
		count = $$1; \
		sample = $$2; \
		total = $$3; \
		totalKB = total / 1024; \
		printf "Total items: %s\n", count; \
		printf "1 item memory usage: %s bytes\n", sample; \
		printf "Total KB used: %s\n", totalKB; \
	}'

setup: infra-up
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Infrastructure is ready!"
