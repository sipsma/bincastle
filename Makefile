SHELL := /bin/bash
export GO111MODULE=on

CACHE_REGISTRY ?= localhost:5000
CACHE_IMAGE_REF ?= buildcache:latest

FUSE_OVERLAYFS_REGISTRY ?= eriksipsma
FUSE_OVERLAYFS_IMAGE_REF ?= bincastle-fuse-overlayfs:latest

BINCASTLE=$(HOME)/.bincastle
BINCASTLE_BIN = $(CURDIR)/bincastle
BINCASTLE_BIN_SRC = $(CURDIR)/cmd/bincastle/bincastle.go
ALL_SRC = $(shell find $(CURDIR) -name '*.go') go.mod go.sum

.PHONY: build
build: $(BINCASTLE_BIN) ;

$(BINCASTLE_BIN): $(ALL_SRC)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -tags "netgo osusergo" -ldflags '-w -extldflags "-static"' -o $(BINCASTLE_BIN) $(BINCASTLE_BIN_SRC)

.PHONY: clean
clean:
	rm -f $(BINCASTLE_BIN)

.PHONY: dist-clean
dist-clean:
	chmod -R u+rwx $(BINCASTLE)/* || true
	rm -rf $(BINCASTLE) || true

.PHONY: fuse-overlayfs
fuse-overlayfs: $(BINCASTLE_BIN)
	$(BINCASTLE_BIN) run --import-cache $(CACHE_REGISTRY)/$(CACHE_IMAGE_REF) --export-image $(FUSE_OVERLAYFS_REGISTRY)/$(FUSE_OVERLAYFS_IMAGE_REF) $(CURDIR) cmd/fuseoverlayfs

.NOTPARALLEL:
