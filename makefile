build-local:
	mkdir -p dist
	go build -o ./dist/bot .

build-run:
	./dist/bot

go-run:
	go run main.go