GO									?= go
NAME								:= $(lastword $(subst /, ,$(CURDIR)))
BUILDFLAGS					:= -v -tags 'netgo'
VERSION 						:= $(shell cat VERSION.txt)
GITCOMMIT 					:= $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES	:= $(shell git status --porcelain --untracked-files=no)
PKG 								:= $(NAME)

ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif
ifeq ($(GITCOMMIT),)
    GITCOMMIT := ${GITHUB_SHA}
endif

CTIMEVAR		?= -X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
LDFLAGS			:= -ldflags "-w $(CTIMEVAR) -extldflags -static"

all: build

.PHONY: build
build:
	@echo "build ${NAME}"
	@$(GO) build $(BUILDFLAGS) $(LDFLAGS) -o $(NAME)

.PHONY: clean
clean:
	@$(GO) clean

