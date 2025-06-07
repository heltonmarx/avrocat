GO					?= go
NAME				:= $(lastword $(subst /, ,$(CURDIR)))
BUILDFLAGS			:= -v -tags 'netgo'
LDFLAGS				:= 
VERSION 			:= $(shell cat VERSION.txt)
GITCOMMIT 			:= $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES	:= $(shell git status --porcelain --untracked-files=no)
PKG 				:= github.meetmecorp.com/hmarques/$(NAME)

ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif
ifeq ($(GITCOMMIT),)
    GITCOMMIT := ${GITHUB_SHA}
endif

CTIMEVAR		?= -X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
LDFLAGS			:= -ldflags "-w $(CTIMEVAR) -extldflags -static"

export CGO_ENABLED	:= 0

all: build

.PHONY: build
build:
	@echo "build ${NAME}"
	@$(GO) build $(BUILDFLAGS) $(LDFLAGS) -o $(NAME)

.PHONY: clean
clean:
	@$(GO) clean

.PHONY: install
install:
	@echo "install schemas"
	@git submodule init && git submodule update

