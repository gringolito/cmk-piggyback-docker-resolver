BINARY=cmk-piggyback-docker-resolver

.PHONY: build test run lint install

build:
	go build -o bin/$(BINARY) ./cmd/$(BINARY)

test:
	go test ./...

run:
	go run ./cmd/$(BINARY) -config ./config.example.yaml

install:
	install -Dm755 bin/$(BINARY) /usr/local/bin/$(BINARY)
	install -Dm644 cmk-piggyback-docker-resolver.service /lib/systemd/system/cmk-piggyback-docker-resolver.service
