.PHONY: test
test: ## Test the app
	go test \
		-v \
		-race \
		-bench=./... \
		-benchmem \
		-timeout=120s \
		-cover \
		-coverprofile=./test_coverage.txt \
		-bench=./... ./...

.PHONY: fmt
fmt: ## Formatting
	gofmt -w -s .
	goimports -w .
	go clean ./...

