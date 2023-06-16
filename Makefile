build:
	docker build -t trevatk/go-chat:v0.0.1 .

deps:
	go mod tidy
	go mod vendor

lint:
	golangci-lint run

rpc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	proto/messenger/v1/messenger_v1.proto
