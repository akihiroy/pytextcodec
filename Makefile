.PHONY: test test-coverage clean lint fmt

# テストを実行
test:
	go test -v ./...

# テストカバレッジを表示
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# リンティングを実行
lint:
	golangci-lint run

# フォーマットを実行
fmt:
	golangci-lint fmt

# クリーンアップ
clean:
	go clean
	rm -f coverage.out
