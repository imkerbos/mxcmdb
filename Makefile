.PHONY: build test lint \
	dev-docker-up dev-docker-up-d dev-docker-down dev-docker-logs dev-docker-restart \
	prod-docker-up prod-docker-up-d prod-docker-down prod-docker-logs prod-docker-restart

# ---------- 构建 & 检查 ----------

build:
	CGO_ENABLED=0 go build -o dist/mxcmdb cmd/server/main.go

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

# ---------- Dev ----------

dev-docker-up:
	docker compose -p mxcmdb -f deploy/dev/docker-compose.yaml up --build

dev-docker-up-d:
	docker compose -p mxcmdb -f deploy/dev/docker-compose.yaml up --build -d

dev-docker-down:
	docker compose -p mxcmdb -f deploy/dev/docker-compose.yaml down

dev-docker-logs:
	docker compose -p mxcmdb -f deploy/dev/docker-compose.yaml logs -f

dev-docker-restart:
	docker compose -p mxcmdb -f deploy/dev/docker-compose.yaml down
	docker compose -p mxcmdb -f deploy/dev/docker-compose.yaml up --build -d

# ---------- Prod ----------

prod-docker-up:
	docker compose -p mxcmdb -f deploy/prod/docker-compose.yaml up --build

prod-docker-up-d:
	docker compose -p mxcmdb -f deploy/prod/docker-compose.yaml up --build -d

prod-docker-down:
	docker compose -p mxcmdb -f deploy/prod/docker-compose.yaml down

prod-docker-logs:
	docker compose -p mxcmdb -f deploy/prod/docker-compose.yaml logs -f

prod-docker-restart:
	docker compose -p mxcmdb -f deploy/prod/docker-compose.yaml down
	docker compose -p mxcmdb -f deploy/prod/docker-compose.yaml up --build -d
