GO ?= go

override GOLANGCI_LINT_FLAGS?=--new-from-rev=HEAD
override GOLANGCI_LINT_FLAGS+=--out-format=colored-line-number


GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
golangci_lint:
	$(call go_get,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,v1.58.1)

GOCTL = $(shell pwd)/bin/goctl
goctl:
	$(call go_get,$(GOCTL),github.com/shippomx/zard/tools/goctl,v1.1.9)

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
lint: golangci_lint pre-test
	@echo ">> linting code..."
	$(GOLANGCI_LINT) $(GOLANGCI_LINT_FLAGS) run

.PHONY: pre-test
pre-test: goctl
	$(GOCTL) template init -home ./test/template
	cp ./model/*.tpl ./test/template/model/
	$(GOCTL) model mysql ddl -dir ./test/model/ -src ./test/sql/*.sql -home ./test/template

.PHONY: after-test
after-test:
	rm -rf ./test/model ./test/template

.PHONY: test
test:
	@go test -race -coverprofile=coverage.tmp.out -coverpkg=./... -covermode=atomic ./...
	@cat coverage.tmp.out \
		| grep -v "gormc/cachedgorm.go" \
		| grep -v "gormc/config/pg/" \
		| grep -v "gormc/sql/" \
		| grep -v "gormc/transition.go" \
		| grep -v "model/main.go" \
		| grep -v "test/" \
		> coverage.out
	@go tool cover -func=coverage.out
	@echo "go test with coverage finished"

.PHONY: test-coverage
test-coverage: pre-test test after-test
