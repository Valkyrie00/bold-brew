##############################
# VARIABLES
##############################
ifneq (,$(wildcard ./.env))
    include .env
    export
endif
%:@

##############################
# CONTAINER
##############################
.PHONY: container-build-image
container-build-image:
	@podman build -f Containerfile -t $(DOCKER_IMAGE_NAME) .

.PHONY: container-build-force-recreate
container-build-force-recreate:
	@podman build --no-cache -f Containerfile -t $(DOCKER_IMAGE_NAME) .

##############################
# RELEASE
##############################
.PHONY: release-snapshot
release-snapshot: container-build-image # Builds the project in snapshot mode and releases it [This is used for testing releases]
	@podman run --rm -v $(PWD):/app $(DOCKER_IMAGE_NAME) goreleaser release --snapshot --clean

.PHONY: build-snapshot # Builds the project in snapshot mode [This is used for testing releases]
build-snapshot: container-build-image
	@podman run --rm -v $(PWD):/app $(DOCKER_IMAGE_NAME) goreleaser build --snapshot --clean

##############################
# BUILD
##############################
.PHONY: build
build: container-build-image
	@podman run --rm -v $(PWD):/app $(DOCKER_IMAGE_NAME) \
	 env GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) go build -o $(APP_NAME) ./cmd/$(APP_NAME)

.PHONY: run
run: build
	./$(APP_NAME)

##############################
# HELPER
##############################
.PHONY: quality
quality: container-build-image
	@podman run --rm -v $(PWD):/app $(DOCKER_IMAGE_NAME) golangci-lint run

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