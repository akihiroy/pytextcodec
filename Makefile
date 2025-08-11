.PHONY: test test-coverage clean lint fmt

# Run tests
test:
	go test -v ./...

# Show test coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run linting
lint:
	golangci-lint run

# Run formatting
fmt:
	golangci-lint fmt

# Clean up
clean:
	go clean
	rm -f coverage.out
