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
		GOOS=darwin GOARCH=amd64 GOPATH=$$(pwd) go $@ $(GO_PROJECT_DIR)/$$Project ; \
		GOOS=linux GOARCH=amd64 GOPATH=$$(pwd) go $@ $(GO_PROJECT_DIR)/$$Project ; \
	done
	@# Make unique binary names for upload to github.
	@for goos in darwin linux ; do \
		if [ -f bin/$${goos}_amd64/grok ] ; then \
			cp -v bin/$${goos}_amd64/grok bin/grok-$${goos}-amd64 ; \
		else \
			cp -v bin/grok bin/grok-$${goos}-amd6 ; \
		fi ; \
	done
	@echo

clean:
	git clean -f -d -x .

bundle:
	$(TAR_EXE) -J -c -f $$(basename $$(pwd))-src.tar.xz \
		--exclude='\.git' \
		$(PROJECTS) \
		test/*.sh test/*.gold test/*.conf \
		Makefile LICENSE README.md
	@ls -l $$(basename $$(pwd))-src.tar.xz

# This is a custom rule for firing up atom on my Mac.
edit:
	@if [[ $$(uname) == "Darwin" ]] ; then \
		echo "installing dlv" ; \
		GOPATH=$$(pwd) go get -u github.com/derekparker/delve/cmd/dlv ; \
	fi
	@echo "starting atom"
	GOPATH=$$(pwd) /opt/atom/latest/Atom.app/Contents/MacOS/Atom

test:
	@cd test && ./test.sh 2>&1 | tee test.log
