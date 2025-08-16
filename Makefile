# ╭────────────────────----------------──────────╮
# │                     go42x                    │
# ╰─────────────────────----------------─────────╯

.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## setup | install dependencies
setup:
	@go mod tidy -e && go mod download
	@go install github.com/go-delve/delve/cmd/dlv@latest
	@go install go.uber.org/mock/mockgen@latest
	@go install github.com/goreleaser/goreleaser/v2@latest
	@go install github.com/anchore/syft/cmd/syft@latest
	@go install github.com/sigstore/cosign/v2/cmd/cosign@latest

# ╭────────────────────----------------──────────╮
# │               General workflow               │
# ╰─────────────────────----------------─────────╯

## test-unit | run unit tests
# -count=1 is needed to prevent caching of test results.
test-unit:
	@go test -count=1 -v -race $(shell go list ./... | grep -v './tests')

## run | run application
# `-N -l` disables compiler optimizations and inlining, which makes debugging easier.
# `[ $$? -eq 1 ]` treats exit code 1 as success. Exit after signal will always be != 0.
run:
	@go run -gcflags="all=-N -l" -race ./main.go $(filter-out $@,$(MAKECMDGOALS)) || [ $$? -eq 1 ]

## run-docker | run application in docker container (linux environment)
# `-N -l` disables compiler optimizations and inlining, which makes debugging easier.
# Using golang image version from go.mod file.
# `[ $$? -eq 1 ]` treats exit code 1 as success. Exit after signal will always be != 0.
run-docker:
	@docker run --rm -it --init \
	-v go-cache:/root/.cache/go-build \
	-v go-mod-cache:/go/pkg/mod \
	-v $(shell pwd):/app \
	-w /app \
	golang:$(shell grep '^go ' go.mod | awk '{print $$2}') \
	go run -gcflags="all=-N -l" -race ./main.go $(filter-out $@,$(MAKECMDGOALS)) || [ $$? -eq 1 ]

## debug | run application with delve debugger
debug:
	@dlv debug ./ --headless --listen=:2345 --accept-multiclient --api-version=2 -- $(filter-out $@,$(MAKECMDGOALS))

## build | build development version of binary
build:
	@go build -gcflags="all=-N -l" -race -v -o ./go42x ./main.go
	@file -h ./go42x && du -h ./go42x && sha256sum ./go42x && go tool buildid ./go42x

## image | build docker image
# @see https://reproducible-builds.org/docs/source-date-epoch/
image:
	@export SOURCE_DATE_EPOCH=0 && \
	docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
    --build-arg "GO_VERSION=$(shell grep '^go ' go.mod | awk '{print $$2}')" \
    --build-arg "COMMIT_HASH=$(shell git rev-parse HEAD 2>/dev/null || echo '')" \
    --build-arg "RELEASE_TAG=$(shell git describe --tags --abbrev=0 2>/dev/null || echo '')" \
	-t ghcr.io/hasansino/go42x:dev \
	.

# ╭────────────────────----------------──────────╮
# │                   Release                    │
# ╰─────────────────────----------------─────────╯

## release-check | validate goreleaser configuration
release-check:
	@goreleaser check

## release-snapshot | build release artifacts without publishing
release-snapshot:
	@goreleaser release --snapshot --clean --skip=publish,sign

## release-local | test the full release process locally
release-local:
	@goreleaser release --skip=publish --clean
