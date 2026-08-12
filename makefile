# FTS5 in mattn/go-sqlite3 requires the sqlite_fts5 build tag.
FTS_TAGS := -tags sqlite_fts5

build-local:
	mkdir -p dist
	go build $(FTS_TAGS) -o ./dist/bot .

build-run:
	./dist/bot

go-run:
	go run $(FTS_TAGS) main.go

test:
	go test $(FTS_TAGS) ./...