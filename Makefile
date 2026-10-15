.PHONY: test report lint
test:
	go test ./...
report:
	go run ./cmd/lab report
lint:
	go vet ./...
	test -z "$$(gofmt -l cmd internal)"
