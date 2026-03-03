include .env
export

DB_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable

SWAGGER_VERSION := v0.30.5
SWAGGER_BIN := $(PWD)/bin/swagger
SWAGGER_YAML := ../swagger.yaml
MODELS_DIR := models
TMP_DIR := tmp_models

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
	OS := linux
endif
ifeq ($(UNAME_S),Darwin)
	OS := darwin
endif

ifeq ($(UNAME_M),x86_64)
	ARCH := amd64
endif
ifeq ($(UNAME_M),aarch64)
	ARCH := arm64
endif
ifeq ($(UNAME_M),arm64)
	ARCH := arm64
endif

.PHONY: generate-models clean

generate-models: $(SWAGGER_BIN)
	@echo "🔧 Generating models into temporary dir '$(TMP_DIR)'..."
	rm -rf $(TMP_DIR)
	mkdir -p $(TMP_DIR)
	$(SWAGGER_BIN) generate model \
		-f $(SWAGGER_YAML) \
		--target $(TMP_DIR)
	@echo "📂 Moving generated files to $(MODELS_DIR)..."
	rm -rf $(MODELS_DIR)
	mkdir -p $(MODELS_DIR)
	if [ -d "$(TMP_DIR)/models" ]; then \
		mv $(TMP_DIR)/models/* $(MODELS_DIR)/; \
	else \
		mv $(TMP_DIR)/* $(MODELS_DIR)/; \
	fi
	rm -rf $(TMP_DIR)
	@echo "🧹 Running go mod tidy..."
	go mod tidy

$(SWAGGER_BIN):
	@echo "⬇️ Installing go-swagger $(SWAGGER_VERSION)..."
	mkdir -p bin
	curl -sSL https://github.com/go-swagger/go-swagger/releases/download/$(SWAGGER_VERSION)/swagger_$(OS)_$(ARCH) -o $(SWAGGER_BIN)
	chmod +x $(SWAGGER_BIN)

clean:
	@echo "🧹 Cleaning up..."
	rm -rf $(SWAGGER_BIN) $(MODELS_DIR) $(TMP_DIR)

migrate.up:
	@echo "🚀 Running database migrations with goose..."
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable" goose -dir migrations -allow-missing -v up

migrate.dictionary.up:
	@echo "🚀 Running dictionary migrations script..."
	go run ./cmd/dictionary/dictionary.go

generate.dictionary:
	@echo "🚀 Running dictionary migrations script..."
	go run ./cmd/dictionary_create_consts/main.go

