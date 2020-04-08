SHELL := /bin/bash
export GO111MODULE=on

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

.NOTPARALLEL:
