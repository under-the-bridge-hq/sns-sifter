.PHONY: build

build: ## sifter CLIをビルド
	go build -o cmd/sifter/sifter ./cmd/sifter/
