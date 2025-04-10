.PHONY: build test clean deps all upload-github create-release

# Build variables
BINARY_NAME=turtlenekko
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# GitHub variables
GITHUB_OWNER=aifoundry-org
GITHUB_REPO=turtlenekko

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get

all: test build

build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/turtlenekko

test:
	$(GOTEST) -v ./...

clean:
	rm -rf $(BUILD_DIR)

deps:
	$(GOMOD) tidy

# Upload binary to GitHub Package Registry
upload-github: build
	@echo "Uploading $(BINARY_NAME) to GitHub Package Registry..."
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "Error: GITHUB_TOKEN environment variable is not set"; \
		exit 1; \
	fi
	@echo "Creating release assets..."
	@mkdir -p $(BUILD_DIR)/release
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)_linux_amd64 ./cmd/turtlenekko
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)_darwin_amd64 ./cmd/turtlenekko
	@echo "Uploading to GitHub..."
	@curl -X POST \
		-H "Authorization: token $(GITHUB_TOKEN)" \
		-H "Accept: application/vnd.github.v3+json" \
		-H "Content-Type: application/octet-stream" \
		--data-binary @$(BUILD_DIR)/release/$(BINARY_NAME)_linux_amd64 \
		"https://uploads.github.com/repos/$(GITHUB_OWNER)/$(GITHUB_REPO)/releases/$(RELEASE_ID)/assets?name=$(BINARY_NAME)_$(VERSION)_linux_amd64"
	@curl -X POST \
		-H "Authorization: token $(GITHUB_TOKEN)" \
		-H "Accept: application/vnd.github.v3+json" \
		-H "Content-Type: application/octet-stream" \
		--data-binary @$(BUILD_DIR)/release/$(BINARY_NAME)_darwin_amd64 \
		"https://uploads.github.com/repos/$(GITHUB_OWNER)/$(GITHUB_REPO)/releases/$(RELEASE_ID)/assets?name=$(BINARY_NAME)_$(VERSION)_darwin_amd64"
	@echo "Upload complete!"

# Create a GitHub release
create-release:
	@echo "Creating GitHub release for version $(VERSION)..."
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "Error: GITHUB_TOKEN environment variable is not set"; \
		exit 1; \
	fi
	@curl -X POST \
		-H "Authorization: token $(GITHUB_TOKEN)" \
		-H "Accept: application/vnd.github.v3+json" \
		-H "Content-Type: application/json" \
		-d '{"tag_name":"v$(VERSION)","name":"v$(VERSION)","body":"Release v$(VERSION)","draft":false,"prerelease":false}' \
		"https://api.github.com/repos/$(GITHUB_OWNER)/$(GITHUB_REPO)/releases" \
		-o $(BUILD_DIR)/release.json
	@echo "Release created! Use RELEASE_ID=$$(cat $(BUILD_DIR)/release.json | jq -r '.id') make upload-github"
