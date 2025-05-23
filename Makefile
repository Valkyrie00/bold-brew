##############################
# VARIABLES
##############################
ifneq (,$(wildcard ./.env))
    include .env
    export
endif
%:@

##############################
# RELEASE
##############################
.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --clean

.PHONY: build-snapshot
build-snapshot:
	goreleaser build --snapshot --clean

##############################
# BUILD
##############################
.PHONY: build
build:
	 @docker run --rm -v $(PWD):/app -w /app golang:$(BUILD_GOVERSION) \
	  env GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) go build -o $(APP_NAME) ./cmd/$(APP_NAME)

.PHONY: run
run: build
	./$(APP_NAME)

##############################
# QUALITY
##############################
.PHONY: lint
lint:
	@golangci-lint run

##############################
# WEBSITE
##############################
.PHONY: build-site
build-site:
	@node build.js

.PHONY: serve-site
serve-site:
	@npx http-server docs -p 3000

.PHONY: dev-site
dev-site: build-site serve-site