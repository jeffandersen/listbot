CMD=promulgate
VERSION?=$(shell git describe --tags --dirty | sed 's/^v//')
GO_BUILD=CGO_ENABLED=0 go build -i --ldflags="-w -X $(shell go list)/version=$(VERSION)"

rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) \
    $(filter $(subst *,%,$2),$d))

LINTERS=\
	gofmt \
	golint \
	vet \
	misspell \
	ineffassign \
	deadcode

all: ci

ci: $(LINTERS) bootstrap vendor build

.PHONY: all ci

# ################################################
# Bootstrapping for base golang package deps
# ################################################

CMD_PKGS=\
	github.com/golang/lint/golint \
	github.com/dominikh/go-tools/simple \
	github.com/client9/misspell/cmd/misspell \
	github.com/gordonklaus/ineffassign \
	github.com/tsenart/deadcode \
	github.com/alecthomas/gometalinter

define VENDOR_BIN_TMPL
vendor/bin/$(notdir $(1)): vendor
	go build -o $$@ ./vendor/$(1)
VENDOR_BINS += vendor/bin/$(notdir $(1))
endef

$(foreach cmd_pkg,$(CMD_PKGS),$(eval $(call VENDOR_BIN_TMPL,$(cmd_pkg))))
$(patsubst %,%-bin,$(filter-out gofmt vet,$(LINTERS))): %-bin: vendor/bin/%
gofmt-bin vet-bin:

bootstrap:
	which dep || go get -u github.com/golang/dep/cmd/dep

vendor: Gopkg.lock
	dep ensure

.PHONY: bootstrap $(CMD_PKGS)

# ################################################
# Test and linting
# ###############################################

$(LINTERS): %: vendor/bin/gometalinter %-bin vendor
	PATH=`pwd`/vendor/bin:$$PATH gometalinter --tests --disable-all --vendor \
	    --deadline=5m -s data ./... --enable $@

.PHONY: $(LINTERS)

# ################################################
# Building
# ###############################################$

LISTBOT_DEPS=\
		vendor

build: $(LISTBOT_DEPS)
	$(GO_BUILD) -o bin/listbot .

heroku: build

.PHONY: build
