docker-build:
	@docker build -t eventdb .

docker-run: docker-build
	@docker run --rm -p 8081:8081 -v .:/usr/src/eventdb eventdb

build:
	@go build -o eventdb