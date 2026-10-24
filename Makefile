.PHONY: run build fmt vet test race check lines
run:
	go run ./cmd/dinosaur
build:
	go build -o bin/dinosaur ./cmd/dinosaur
fmt:
	go fmt ./...
vet:
	go vet ./...
test:
	go test ./...
race:
	go test -race ./...
lines:
	python3 scripts/check_file_lines.py
check: lines vet test
