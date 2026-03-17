APP = ia
CMD_DIR = ./cmd/ia
BIN_DIR = ./bin
BIN = $(BIN_DIR)/$(APP)
INSTALL_DIR ?= ~/.local/bin

.PHONY: help build install run dry-run clean
.DEFAULT_GOAL := help

help:
	@echo "Usage:"
	@echo "  make build"
	@echo "  make install"
	@echo "  make run ARGS='<agent> <language> <project> [--dry-run]'"
	@echo "  make dry-run ARGS='<agent> <language> <project>'"
	@echo "  make clean"
	@echo "  make lint"

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(CMD_DIR)

install: build
	install -m 0755 $(BIN) $(INSTALL_DIR)/$(APP)

run:
	go run $(CMD_DIR) $(ARGS)

dry-run:
	go run $(CMD_DIR) $(ARGS) --dry-run

clean:
	rm -rf $(BIN_DIR)

lint:
	golangci-lint run
