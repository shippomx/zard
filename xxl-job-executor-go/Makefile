GO ?= go

override GOLANGCI_LINT_FLAGS?=--new-from-rev=HEAD
override GOLANGCI_LINT_FLAGS+=--out-format=colored-line-number



GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
golangci_lint:
	$(call go_get,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,v1.59.0)

# go_get will 'go get' any package $2@$3 and install it to $1.
# usage $(call go_get,$BinaryLocalPath,$GoModuleName,$Version)
define go_get
@set -e; \
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


.PHONY: lint
lint: check-path golangci_lint
	@echo ">> linting code..."
	$(GOLANGCI_LINT) $(GOLANGCI_LINT_FLAGS) run 

check-path:
	@echo ">> checking GOPATH..."
	@echo "GOPATH: $(GOPATH)"
	@echo "Current dir: $(PWD)"
	@echo "Package path: $(shell go list -m)"

.PHONY: test
test:
	@go test -race ./...
	@echo "go test finished"

.PHONY: test-coverage
test-coverage:
	@go list ./...  | grep -vE '/(example)($$|/)' | xargs go test -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out
	@echo "go test with coverage finished"


