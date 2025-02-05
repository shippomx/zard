GO ?= go

override GOLANGCI_LINT_FLAGS?=--new-from-rev=HEAD
override GOLANGCI_LINT_FLAGS+=--out-format=colored-line-number

VERSION_TAG?=$(shell git describe --tags --abbrev=0)

LINT_LOCAL_DOCKER?=true

DOCKER_IMAGE := nexus-dev-image.fulltrust.link/base-images/golang:1.22-lint
# 仅指定本地使用，docker自带
GOLANGCI_LINT = $(shell if [ -f $(GOPATH)/bin/golangci-lint ]; then echo $(GOPATH)/bin/golangci-lint; else echo $(shell pwd)/bin/golangci-lint; fi)
# go_get will 'go get' any package $2@$3 and install it to $1.
# usage $(call go_get,$BinaryLocalPath,$GoModuleName,$Version)
define go_get
set -e; \
if [ -f ${1} ]; then \
	[ -z ${3} ] && exit 0; \
	installed_version=$$(go version -m "${1}" | grep -E '[[:space:]]+mod[[:space:]]+' | awk '{print $$3}') ; \
	[ "$${installed_version}" = "${3}" ] && exit 0; \
	echo ">> ${1} ${2} $${installed_version}!=${3}, ${3} will be installed."; \
fi; \
module=${2}; \
if ! [ -z ${3} ]; then module=${2}@${3}; fi; \
echo "Downloading $${module}" ;\
GOBIN=$(shell pwd)/bin $(GO) install $${module} ;
endef

.PHONY: go-env-init
go-env-init:
	go env -w GO111MODULE=on

.PHONY: lint
lint: git-config-init go-env-init
	@echo ">> linting code..."
	@if [ -f "/proc/1/cgroup" ] && grep -q "docker" "/proc/1/cgroup" || [ -f "/.dockerenv" ] ; then \
		echo "run in online docker"; \
		golangci-lint $(GOLANGCI_LINT_FLAGS) run; \
	else \
		if command -v docker > /dev/null 2>&1 && [ $(LINT_LOCAL_DOCKER) = "true" ] ; then \
			echo "run in local docker"; \
			docker run --rm \
				-v $(shell pwd):$(shell pwd) \
				-v ~/.ssh:/root/.ssh \
				-v ~/.gitconfig:/root/.gitconfig \
				-w $(shell pwd) $(DOCKER_IMAGE) \
				sh -c "go mod tidy && golangci-lint $(GOLANGCI_LINT_FLAGS) run"; \
		else \
			echo "docker is not installed or specified to run locally"; \
			$(call go_get,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,v1.60.1) \
			$(GOLANGCI_LINT) $(GOLANGCI_LINT_FLAGS) run; \
		fi \
	fi

.PHONY: test-goctl
test-goctl: git-config-init go-env-init
	./hack/test-goctl.sh test-goctl::run

.PHONY: test
test:
	@go test -race ./...
	@echo "go test finished"

.PHONY: test-coverage
test-coverage: git-config-init go-env-init
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out
	@echo "go test with coverage finished"

.PHONY: release-tag
release:
	@echo $(VERSION_TAG)
	@sed -i '' -e "s/const BuildVersion = \".*\"/const BuildVersion = \"$(VERSION_TAG)\"/" core/utils/version.go
	@sed -i '' -e "s/const BuildVersion = \".*\"/const BuildVersion = \"$(VERSION_TAG)\"/" tools/goctl/internal/version/version.go
	@git diff -p
