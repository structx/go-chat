
build:
	docker build -t trevatk/messenger:v0.1.0 .

rpc:
	protoc -I=../ proto/messenger/v1/messenger_v1.proto \
  	--js_out=import_style=commonjs,binary:. \
  	--grpc-web_out=import_style=typescript,mode=grpcweb:.