.PHONY: default
default: test

include common.mk

.PHONY: test
test: go-test-all

.PHONY: lint
lint: go-lint-all helm-lint git-clean-check

.PHONY: generate
generate: buf-generate-all typescript-compile

.PHONY: build-server
build-server:
	go build -o ./bin/server ./server/cmd/

.PHONY: build-loader
build-loader:
	go build -o ./bin/loader ./loader/cmd/

.PHONY: build-docker-server
build-docker-server:
	docker build --build-arg TARGETARCH=amd64 -t llmariner/model-manager-server:latest -f build/server/Dockerfile .

.PHONY: build-docker-loader
build-docker-loader:
	docker build --build-arg TARGETARCH=amd64 -t llmariner/model-manager-loader:latest -f build/loader/Dockerfile .

.PHONY: build-docker-convert-gguf
build-docker-convert-gguf:
	docker build --build-arg TARGETARCH=amd64 -t llmariner/experiments-convert_gguf:latest -f build/experiments/convert_gguf/Dockerfile build/experiments/convert_gguf/

.PHONY: check-helm-tool
check-helm-tool:
	@command -v helm-tool >/dev/null 2>&1 || $(MAKE) install-helm-tool

.PHONY: install-helm-tool
install-helm-tool:
	go install github.com/cert-manager/helm-tool@latest

.PHONY: generate-chart-schema
generate-chart-schema: generate-chart-schema-server generate-chart-schema-loader

.PHONY: generate-chart-schema-server
generate-chart-schema-server: check-helm-tool
	@cd ./deployments/server && helm-tool schema > values.schema.json

.PHONY: generate-chart-schema-loader
generate-chart-schema-loader: check-helm-tool
	@cd ./deployments/loader && helm-tool schema > values.schema.json

.PHONY: helm-lint
helm-lint: helm-lint-server helm-lint-loader

.PHONY: helm-lint-server
helm-lint-server: generate-chart-schema-server
	cd ./deployments/server && helm-tool lint
	helm lint ./deployments/server

.PHONY: helm-lint-loader
helm-lint-loader: generate-chart-schema-loader
	cd ./deployments/loader && helm-tool lint
	helm lint ./deployments/loader
