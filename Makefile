TARGET=api
default: $(TARGET)

PORT=8080

.PHONY: run
run: ## Start a local API server instance
	@go run cmd/api/api.go

.PHONY: clean
clean: ## Clean most generated files
	@rm -rf \
		$(TARGET) \
		*.out \
		*.log

.PHONY: help
help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'
