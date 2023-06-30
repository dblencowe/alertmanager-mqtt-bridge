build:
	go build ./cmd/main.go

test:
	go test ./...

docker-build:
	docker build -t alertmanager-mqtt-bridge:latest -f Dockerfile .