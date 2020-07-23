SHELL := /bin/bash
export GO111MODULE=on

CACHE_REGISTRY ?= localhost:5000
CACHE_IMAGE_REF ?= buildcache:latest

FUSE_OVERLAYFS_REGISTRY ?= localhost:5000
FUSE_OVERLAYFS_IMAGE_REF ?= fuse-overlayfs:latest

BINCASTLE=$(HOME)/.bincastle
BINCASTLE_BIN = $(CURDIR)/bincastle
BINCASTLE_BIN_SRC = $(CURDIR)/cmd/bincastle/bincastle.go
ALL_SRC = $(wildcard $(CURDIR)/**/*.go)
$(BINCASTLE_BIN): $(ALL_SRC)
	rm $(BINCASTLE_BIN) || true
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -tags "netgo osusergo" -ldflags '-w -extldflags "-static"' -o $(BINCASTLE_BIN) $(BINCASTLE_BIN_SRC)

.PHONY: clean-bin
clean-bin:
	rm $(BINCASTLE_BIN) || true

.PHONY: clean-state
clean-state:
	chmod -R u+rwx $(BINCASTLE)/* || true
	rm -rf $(BINCASTLE)/* || true

.PHONY: rebuild
rebuild: clean-bin $(BINCASTLE_BIN)

.PHONY: fuse-overlayfs
fuse-overlayfs: $(BINCASTLE_BIN)
	$(BINCASTLE_BIN) run --import-cache $(CACHE_REGISTRY)/$(CACHE_IMAGE_REF) --export-image $(FUSE_OVERLAYFS_REGISTRY)/$(FUSE_OVERLAYFS_IMAGE_REF) $(CURDIR) cmd/fuseoverlayfs

.NOTPARALLEL:
