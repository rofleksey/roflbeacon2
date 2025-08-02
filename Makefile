.PHONY: all
all: gen build run

.PHONY: clean
clean:
	@go clean

.PHONY: gen
gen:
	@echo "Generating dependency files..."
	@go generate ./...

.PHONY: lint
lint:
	@npx @redocly/cli lint
	@npx golangci-lint run

.PHONY: docs
docs:
	@npx @redocly/cli build-docs --title "RoflBeacon2 API" -o docs/api.html app/api/api/openapi.yaml

.PHONY: build
build:
	@echo "Building application..."
	@go build -ldflags "-X roflbeacon2/pkg/build.Tag=${GIT_TAG}" .

.PHONY: run
run:
	@./roflbeacon2
