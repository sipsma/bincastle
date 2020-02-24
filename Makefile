SHELL := /bin/bash
export GO111MODULE=on

BINCASTLE=$(HOME)/.bincastle
EXAMPLE_BIN_NAME = system
EXAMPLE_BIN = $(CURDIR)/$(EXAMPLE_BIN_NAME)
EXAMPLE_SRC = $(CURDIR)/system.go
ALL_SRC = $(wildcard $(CURDIR)/**/*.go)
$(EXAMPLE_BIN): $(ALL_SRC)
	rm $(EXAMPLE_BIN) || true
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -tags "netgo osusergo" -ldflags '-w -extldflags "-static"' -o $(EXAMPLE_BIN) $(EXAMPLE_SRC)

.PHONY: run-system
run-system: $(EXAMPLE_BIN)
	$(EXAMPLE_BIN) run system

.PHONY: export-system
export-system: $(EXAMPLE_BIN)
	$(EXAMPLE_BIN) export localhost:5000/system

.PHONY: clean-bin
clean-bin:
	rm $(EXAMPLE_BIN) || true

.PHONY: clean-state
clean-state:
	chmod -R u+rwx $(BINCASTLE)/* || true
	rm -rf $(BINCASTLE)/* || true

.PHONY: rebuild
rebuild: clean-bin $(EXAMPLE_BIN)

.PHONY: rerun-system
rerun-system: clean-bin run-system

.PHONY: reexport-system
reexport-system: clean-bin export-system

.NOTPARALLEL:
