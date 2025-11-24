# Makefile
test:
	go test ./... -v -race -cover
run:
	go run cmd/main.go
