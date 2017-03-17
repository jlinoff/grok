# ================================================================
# Makefile for the grok project.
# ================================================================
GO_PROJECT_DIR := jlinoff
PROJECTS       := $(wildcard src/$(GO_PROJECT_DIR)/*)
GO_PROJECTS    := $(patsubst src/$(GO_PROJECT_DIR)/%, %, $(PROJECTS))
TAR_EXE        := $(shell which gnutar)

ifeq "$(TAR_EXE)" ""
    TAR_EXE := $(shell which tar)
endif

# ================================================================
# Rules.
# ================================================================
.PHONY: all bundle clean edit install test

all: install test

install:
	@for Project in $(GO_PROJECTS) ; do \
		echo "" ; \
		echo "GOPATH=$$(pwd) go $@ $(GO_PROJECT_DIR)/$$Project" ; \
		GOPATH=$$(pwd) go $@ $(GO_PROJECT_DIR)/$$Project ; \
	done
	@echo

clean:
	@for Project in $(GO_PROJECTS) ; do \
		echo "GOPATH=$$(pwd) go $@ $(GO_PROJECT_DIR)/$$Project" ; \
		GOPATH=$$(pwd) go $@ $(GO_PROJECT_DIR)/$$Project ; \
	done
	rm -rf test/*/test *log
	find . -name '*~' -delete
	find . -name '*log' -delete

bundle:
	$(TAR_EXE) -J -c -f $$(basename $$(pwd))-src.tar.xz \
		--exclude='\.git' \
		$(PROJECTS) \
		test/*.sh test/*.gold test/*.conf \
		Makefile LICENSE README.md
	@ls -l $$(basename $$(pwd))-src.tar.xz

edit:
	GOPATH=$$(pwd) /opt/atom/latest/Atom.app/Contents/MacOS/Atom

test:
	@cd test && ./test.sh 2>&1 | tee test.log

