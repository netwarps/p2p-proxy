#
# list make targets https://gist.github.com/pvdb/777954
#
# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

# We don't need make's built-in rules.
.SUFFIXES:
MAKEFLAGS += --no-builtin-rules

# Constants used throughout.
.EXPORT_ALL_VARIABLES:
OUT_DIR ?= _output
BIN_DIR := $(OUT_DIR)/bin
BINARY ?= p2p-proxy
VERSION ?= v0.3.0
COMMIT_SHA = $(shell git rev-parse --short HEAD)

MODULE = $(shell go list -m)
GO_SOURCE_FILES := $(shell find . -type f -name '*.go')
GO_TAGS ?=
#GO_LDFLAGS ?= "-s -w"
GO_LDFLAGS ?=

# defined in metadata/metadata.go
METADATA_VAR = Version=$(VERSION)
METADATA_VAR += CommitSHA=$(COMMIT_SHA)

RELEASE_PLATFORMS = darwin-amd64 linux-amd64 windows-amd64

extensions.windows = .exe

.PHONY: build
build: $(OUT_DIR)/bin/$(BINARY)

$(OUT_DIR)/bin/$(BINARY): GO_LDFLAGS += $(METADATA_VAR:%=-X $(MODULE)/metadata.%)
$(OUT_DIR)/bin/$(BINARY): $(GO_SOURCE_FILES)
	@echo "Building $@"
	@mkdir -p $(@D)
	GOBIN=$(abspath $(@D)) go install -tags "$(GO_TAGS)" -ldflags "-s -w $(GO_LDFLAGS)" $(MODULE)
	@touch $@

# builds release packages for all target platforms
.PHONY: release
release: $(RELEASE_PLATFORMS:%=%)

.PHONY: $(RELEASE_PLATFORMS:%=%)
$(RELEASE_PLATFORMS:%=%): GO_LDFLAGS += $(METADATA_VAR:%=-X $(MODULE)/metadata.%)
$(RELEASE_PLATFORMS:%=%): %: $(OUT_DIR)/%/bin/$(BINARY) README.md
	$(eval GOOS = $(word 1,$(subst -, ,$(@F))))
	@mkdir -p $(OUT_DIR)/$(@F)/release
	@cp $(OUT_DIR)/$(@F)/bin/* $(OUT_DIR)/$(@F)/release/
	@cp README.md $(OUT_DIR)/$(@F)/release/ 
	@[ "$(GOOS)" == "windows" ] && cd $(OUT_DIR)/$(@F)/release && zip $(BINARY)-$(@F).$(VERSION).zip * || true
	@[ "$(GOOS)" != "windows" ] && cd $(OUT_DIR)/$(@F)/release && tar -czvf $(BINARY)-$(@F).$(VERSION).tar.gz * || true


# explicit targets for all platform executables
$(RELEASE_PLATFORMS:%=$(OUT_DIR)/%/bin/$(BINARY)): $(GO_SOURCE_FILES)
	$(eval platform = $(patsubst $(OUT_DIR)/%/bin,%,$(@D)))
	$(eval GOOS = $(word 1,$(subst -, ,$(platform))))
	$(eval GOARCH = $(word 2,$(subst -, ,$(platform))))
	@echo "Building $@ for $(GOOS)-$(GOARCH)"
	mkdir -p $(@D)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@$(extensions.$(GOOS)) -tags "$(GO_TAGS)" -ldflags "-s -w $(GO_LDFLAGS)" $(MODULE)
	
.PHONY: clean
clean:
	@rm -rf $(OUT_DIR)
docker:
	docker build -t registry.paradeum.com/netwarps/p2p-proxy:v0.0.1 .