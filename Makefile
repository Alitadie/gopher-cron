# Makefile
test:
	go test ./... -v -race -cover
run:
	go run cmd/main.go
# 如果不想本地安装 protoc，可以用 docker
proto:
	docker run --rm -v $(PWD):/work uber/prototool:latest protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/scheduler.proto
