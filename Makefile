VERSION=1.17.1
FILE_COMMAND=terragrunt-atlantis-config
DIR_BUILD=./build
PATH_BUILD=${DIR_BUILD}/${FILE_COMMAND}/${VERSION}
PATTERN_BUILD=${PATH_BUILD}/${FILE_COMMAND}_${VERSION}_{{.OS}}_{{.Arch}}

SHELL = bash

ifeq ($(OS),Windows_NT)
	UNAME=windows
else
	UNAME=$(shell uname | tr '[:upper:]' '[:lower:]')
endif

ARCH = amd64
FILE_ARCH=$(UNAME)_$(ARCH)

S3_BUCKET_NAME=cloudfront-origin-homebrew-tap-transcend-io
PROFILE=transcend-prod

# Determine the arch/os combos we're building for
XC_OSARCH=linux/amd64 linux/arm darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 linux/arm64

ARCHIVE_FILES = README.md

.PHONY: clean
clean:
	rm -rf ${DIR_BUILD}
	rm -rf '$(HOME)/bin/$(FILE_COMMAND)'

.PHONY: build
build: clean
	gox \
		-os="$(UNAME)" \
		-arch="$(ARCH)" \
		-output="$(PATTERN_BUILD)" \
		-ldflags "-X main.VERSION=$(VERSION)"

.PHONY: build-all
build-all: clean
	gox \
		-osarch="$(XC_OSARCH)" \
		-output="$(PATTERN_BUILD)" \
		-ldflags "-X main.VERSION=$(VERSION)"

.PHONY: gotestsum
gotestsum:
	mkdir -p cmd/test_artifacts
	gotestsum
	rm -rf cmd/test_artifacts

.PHONY: test
test:
	mkdir -p cmd/test_artifacts
	go test -v ./...
	rm -rf cmd/test_artifacts

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: sign
sign: archives
	rm -f ${DIR_BUILD}/$(VERSION)/SHA256SUMS
	cd ${DIR_BUILD}/${VERSION} && shasum -a256 * > SHA256SUMS

.PHONY: install
install:
	install -d -m 755 '$(HOME)/bin/'
	install $(PATH_BUILD)/$(FILE_COMMAND)_$(VERSION)_$(FILE_ARCH) '$(HOME)/bin/$(FILE_COMMAND)'

.PHONY: archives
archives: build-all
	mkdir -p build/$(VERSION) build/tmp

	for file in $(PATH_BUILD)/*; do \
		base=$$(basename $$file .exe); \
		case $$file in \
			*darwin* | *windows*) \
				mkdir -p ${DIR_BUILD}/tmp/$$base; \
				cp $(ARCHIVE_FILES) $$file ${DIR_BUILD}/tmp/$$base; \
				cd ${DIR_BUILD}/tmp && zip -r ../$(VERSION)/$$base.zip $$base && cd ../..; \
				rm -rf ${DIR_BUILD}/tmp/$$base; \
				;; \
			*) \
				tar -zcvf ./build/$(VERSION)/$$base.tar.gz --transform "s/^/$$base\//" -C . $(ARCHIVE_FILES) -C $$(dirname $$file) $$base; \
			;; \
		esac \
	done \

	rm -rf ${DIR_BUILD}/tmp
