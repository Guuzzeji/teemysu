# FTS5 in mattn/go-sqlite3 requires the sqlite_fts5 build tag.
FTS_TAGS := -tags sqlite_fts5

build-local:
	mkdir -p dist
	go build $(FTS_TAGS) -o ./dist/bot .

build-run:
	./dist/bot

go-run:
	go run $(FTS_TAGS) main.go

# Run the built binary (bot loads .env itself via godotenv).
run-local: build-local
	./dist/bot

test:
	go test $(FTS_TAGS) ./...

# Docker: build image + run via compose (needs .env, cp .env.example .env first).
docker-build:
	docker build -t teemysu:latest .

docker-up:
	docker compose --env-file .env up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f