.PHONY: run build frontend fmt vet test race check lines
frontend:
	./build.sh --frontend-only
run: frontend
	go run ./cmd/dinosaur
build:
	./build.sh
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
check: frontend
	python3 scripts/check_file_lines.py
	go vet ./...
	go test ./...
