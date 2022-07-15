# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.2.1. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Bellow generated variables ensure that every time a tool under each variable is invoked, the correct version
# will be used; reinstalling only if needed.
# For example for bingo variable:
#
# In your main Makefile (for non array binaries):
#
#include .bingo/Variables.mk # Assuming -dir was set to .bingo .
#
#command: $(BINGO)
#	@echo "Running bingo"
#	@$(BINGO) <flags/args..>
#
BINGO := $(GOBIN)/bingo-v0.2.2
$(BINGO): .bingo/bingo.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/bingo-v0.2.2"
	@cd .bingo && $(GO) build -modfile=bingo.mod -o=$(GOBIN)/bingo-v0.2.2 "github.com/bwplotka/bingo"

EMBEDMD := $(GOBIN)/embedmd-v1.0.0
$(EMBEDMD): .bingo/embedmd.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/embedmd-v1.0.0"
	@cd .bingo && $(GO) build -modfile=embedmd.mod -o=$(GOBIN)/embedmd-v1.0.0 "github.com/campoy/embedmd"

GOTEST := $(GOBIN)/gotest-v0.0.4
$(GOTEST): .bingo/gotest.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/gotest-v0.0.4"
	@cd .bingo && $(GO) build -modfile=gotest.mod -o=$(GOBIN)/gotest-v0.0.4 "github.com/rakyll/gotest"

LICHE := $(GOBIN)/liche-v0.0.0-20200229003944-f57a5d1c5be4
$(LICHE): .bingo/liche.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/liche-v0.0.0-20200229003944-f57a5d1c5be4"
	@cd .bingo && $(GO) build -modfile=liche.mod -o=$(GOBIN)/liche-v0.0.0-20200229003944-f57a5d1c5be4 "github.com/raviqqe/liche"

