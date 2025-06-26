##############################
# VARIABLES
##############################
ifneq (,$(wildcard ./.env))
    include .env
    export
endif
%:@

##############################
# DOCKER
##############################
.PHONY: docker-build-image
docker-build-image:
	@docker build -t $(DOCKER_IMAGE_NAME) .

docker-build-force-recreate:
	@docker build --no-cache -t $(DOCKER_IMAGE_NAME) .

##############################
# RELEASE
##############################
.PHONY: release-snapshot
release-snapshot: docker-build-image # Builds the project in snapshot mode and releases it [This is used for testing releases]
	@docker run --rm -v $(PWD):/app $(DOCKER_IMAGE_NAME) goreleaser release --snapshot --clean

.PHONY: build-snapshot # Builds the project in snapshot mode [This is used for testing releases]
build-snapshot: docker-build-image
	@docker run --rm -v $(PWD):/app $(DOCKER_IMAGE_NAME) goreleaser build --snapshot --clean

##############################
# BUILD
##############################
.PHONY: build
build: docker-build-image
	@docker run --rm -v $(PWD):/app $(DOCKER_IMAGE_NAME) \
	 env GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) go build -o $(APP_NAME) ./cmd/$(APP_NAME)

.PHONY: run
run: build
	./$(APP_NAME)

##############################
# HELPER
##############################
.PHONY: quality
quality: docker-build-image
	@docker run --rm -v $(PWD):/app $(DOCKER_IMAGE_NAME) golangci-lint run

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