.PHONY: setup server help build

VENV := xmcp/.venv
PYTHON := $(VENV)/bin/python

help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

setup: $(VENV)/bin/activate ## Python仮想環境を作成し依存パッケージをインストール

$(VENV)/bin/activate: xmcp/requirements.txt
	python3 -m venv $(VENV)
	$(VENV)/bin/pip install -r xmcp/requirements.txt
	touch $(VENV)/bin/activate

xmcp/.env:
	@if [ ! -f xmcp/.env ]; then \
		cp xmcp/.env.example.local xmcp/.env; \
		echo "xmcp/.env を作成しました。認証情報を設定してください。"; \
	fi

build: ## sifter CLIをビルド
	go build -o cmd/sifter/sifter ./cmd/sifter/

server: setup xmcp/.env ## xmcp MCPサーバーを起動
	$(PYTHON) xmcp/server.py
