GO ?= go

override GOLANGCI_LINT_FLAGS?=--new-from-rev=HEAD
override GOLANGCI_LINT_FLAGS+=--out-format=colored-line-number

LINT_LOCAL_DOCKER?=true

DOCKER_IMAGE := nexus-dev-image.fulltrust.link/base-images/golang:1.22-lint

GOCTL_LOCAL_GOZERO :=
ifeq ($(IF_TEST_GOCTL), true)
    GOCTL_LOCAL_GOZERO = -v $(shell pwd)/../../:$(shell pwd)/../../
endif

GOLANGCI_LINT = $(shell if [ -f $(GOPATH)/bin/golangci-lint ]; then echo $(GOPATH)/bin/golangci-lint; else echo $(shell pwd)/bin/golangci-lint; fi)

.PHONY: lint
lint:
	@echo ">> linting code..."
	@if [ -f "/proc/1/cgroup" ] && grep -q "docker" "/proc/1/cgroup" || [ -f "/.dockerenv" ] ; then \
		echo "run in online docker"; \
		golangci-lint $(GOLANGCI_LINT_FLAGS) run; \
	else \
		if command -v docker > /dev/null 2>&1 && [ $(LINT_LOCAL_DOCKER) = "true" ] ; then \
			echo "run in local docker"; \
			docker run --rm \
				-v $(shell pwd):$(shell pwd) \
				$(GOCTL_LOCAL_GOZERO)  \
				-v ~/.ssh:/root/.ssh \
				-v ~/.gitconfig:/root/.gitconfig \
				-w $(shell pwd) $(DOCKER_IMAGE) \
				sh -c "go mod tidy && golangci-lint $(GOLANGCI_LINT_FLAGS) run"; \
		else \
			echo "docker is not installed or specified to run locally"; \
			$(GOLANGCI_LINT) $(GOLANGCI_LINT_FLAGS) run; \
		fi \
	fi

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

clean:
	# Clean up the compilation product
	rm -rf out

prepare:
	mkdir -p out
	# download 3rd package
	go mod download

build: clean prepare
	go build -o out