.DEFAULT_GOAL := dev

.PHONY: dev
dev: ## dev build
dev: clean install generate buildweb vet fmt lint test mod-tidy

.PHONY: ci
ci: ## CI build
ci: dev diff

.PHONY: clean
clean: ## remove files created during build pipeline
	$(call print-target)
	rm -rf web/app/node_modules
	rm -rf web/app/public/build
	rm -rf dist
	rm -f coverage.*
	rm -rf .regodeps

.PHONY: install
install: ## go install tools
	$(call print-target)
	cd tools && go install $(shell cd tools && go list -f '{{ join .Imports " " }}' -tags=tools)

.PHONY: generate
generate: ## go generate
	$(call print-target)
	go generate ./...

.PHONY: buildweb
buildweb: ## generate webapp build artifacts
	$(call print-target)
	cd web/app && yarn && yarn build

.PHONY: vet
vet: ## go vet
	$(call print-target)
	go vet ./...

.PHONY: fmt
fmt: ## go fmt and opa
	$(call print-target)
	go fmt ./...
	opa fmt -w .
	cd web/app && yarn format

.PHONY: lint
lint: ## golangci-lint and opa
	$(call print-target)
	golangci-lint run
	opa fmt --fail -l .
	go run ./devtools/regodeps
	opa check --strict `find . -type f -name "*.rego"`
	cd web/app && yarn run svelte-check && yarn lint

.PHONY: test
test: ## go test with race detector and code covarage
	$(call print-target)
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: mod-tidy
mod-tidy: ## go mod tidy
	$(call print-target)
	go mod tidy
	cd tools && go mod tidy

.PHONY: diff
diff: ## git diff
	$(call print-target)
	git diff --exit-code
	RES=$$(git status --porcelain) ; if [ -n "$$RES" ]; then echo $$RES && exit 1 ; fi

.PHONY: build
build: ## goreleaser --snapshot --skip-publish --clean
build: install buildweb
	$(call print-target)
	goreleaser --snapshot --skip-publish --clean

.PHONY: release
release: ## goreleaser --clean
release: install buildweb
	$(call print-target)
	goreleaser --clean

.PHONY: run-live
run-live: ## go run
	@go run -race ./cmd/pslive

.PHONY: run-local
run-local: ## go run
	@go run -race ./cmd/pslocal

.PHONY: go-clean
go-clean: ## go clean build, test and modules caches
	$(call print-target)
	go clean -r -i -cache -testcache -modcache

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

define print-target
    @printf "Executing target: \033[36m$@\033[0m\n"
endef
