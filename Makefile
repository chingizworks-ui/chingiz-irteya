BUILD_DATE = $(shell date +%D-%H:%M)

ifeq ($(CI), true)
	COMMIT_HASH = $(CI_COMMIT_SHORT_SHA)
	COMMIT_DATE = $(CI_COMMIT_TIMESTAMP)
	ifeq ($(CI_PIPELINE_SOURCE), merge_request_event)
		BUILD_BRANCH = $(CI_MERGE_REQUEST_SOURCE_BRANCH_NAME)
	else
		BUILD_BRANCH = $(CI_COMMIT_BRANCH)
		BUILD_VERSION = $(CI_COMMIT_TAG)
	endif
else
	BUILD_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
	BUILD_VERSION = $(shell git describe --tags --abbrev=0 HEAD 2>/dev/null)
	COMMIT_HASH = $(shell git rev-parse --short HEAD)
	COMMIT_DATE = $(shell git show --no-patch --format=%ci)
endif

ifeq ($(BUILD_VERSION),)
	BUILD_VERSION = $(BUILD_BRANCH)-$(COMMIT_HASH)
endif

LDFLAGS = \
	-s \
	-w \
	-X 'stockpilot/internal/app._Branch=$(BUILD_BRANCH)' \
	-X 'stockpilot/internal/app._BuildVersion=$(BUILD_VERSION)' \
	-X 'stockpilot/internal/app._CommitHash=$(COMMIT_HASH)' \
	-X 'stockpilot/internal/app._CommitDate=$(COMMIT_DATE)' \
	-X 'stockpilot/internal/app._BuildDate=$(BUILD_DATE)'

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o binaries/stockpilot-service ./cmd/api

build_native:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -a -installsuffix cgo -o binaries/stockpilot-service-native ./cmd/api

run_tests:
	go test ./... -v

test: build_native run_tests

swagger:
	go install github.com/swaggo/swag/cmd/swag@v1.16.4
	swag init --parseDependency -g cmd/api/main.go -o docs

up:
	docker-compose up --build

deps-up:
	docker-compose up -d db otel-collector

migrate:
	psql $$DATABASE_URL -f migrations/0001_init.sql
