APP_NAME = api
CMD_PATH = ./cmd/$(APP_NAME)
BUILD_DIR = ./bin

GOFLAGS = -trimpath -ldflags "-s -w"
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: all build clean run

build:
	@echo "🔨 Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "✅ Build completed: $(BUILD_DIR)/$(APP_NAME)"

build-check:
	@echo "🔨 Running build check $(APP_NAME)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GOFLAGS) -o /dev/null $(CMD_PATH)
	@echo "✅ Build ok"

run: build
	@echo "🚀 Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

format:
	@echo "🚀 Formatting $(APP_NAME)..."
	go fmt ./...

format-check:
	@echo "🚀 Running format check $(APP_NAME)..."
	./scripts/fmt.sh
	@echo "✅ Format ok"

install-staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

test:
	@echo "🚀 Running test $(APP_NAME)..."
	go test ./... -timeout 30s

test-single:
	@echo "🚀 Running test $(TEST_NAME) in $(APP_NAME)..."
	go test -run $(TEST_NAME) ./... -v -timeout 30s

lint:
	@echo "🚀 Running linter $(APP_NAME)..."
	go vet ./...
	~/go/bin/staticcheck ./...
	@echo "✅ Lint ok"

run-watch:
	@echo "🚀 Running $(APP_NAME)..."
	~/go/bin/air -c .air.toml

clean:
	@echo "🧹 Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

ci: install-staticcheck format-check lint build-check test

# Check dockerfile
dockerfile-check:
	@echo "🚀 Running dockerfile from scratch..."
	docker compose down --rmi all --volumes --remove-orphans
	docker compose build --no-cache
	docker compose up

dockerfile-up:
	@echo "🚀 Running dockerfile..."
	docker compose up --build
