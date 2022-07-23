
GIT_VERSION ?= $(shell git describe --tags --always --dirty 2> /dev/null)
VERSION = $(shell echo $(GIT_VERSION) | sed 's/^v//' | sed 's/-.*//')

PROVIDER_HOSTNAME=github.com
PROVIDER_NAMESPACE=tensor5
PROVIDER_TYPE=talos
PROVIDER_TARGET=$(shell go env GOOS)_$(shell go env GOARCH)
PROVIDER_PATH=~/.terraform.d/plugins/$(PROVIDER_HOSTNAME)/$(PROVIDER_NAMESPACE)/$(PROVIDER_TYPE)/$(VERSION)/$(PROVIDER_TARGET)

default: testacc

build:
	@mkdir -p $(PROVIDER_PATH)
	go build \
		-tags release \
		-ldflags '-X $(MODULE)/internal/config.Version=$(GIT_VERSION)' \
		-o $(PROVIDER_PATH)/terraform-provider-$(PROVIDER_TYPE)_v$(VERSION)

.PHONY: docs
docs:
	go generate

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
