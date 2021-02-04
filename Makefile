VERSION=1.1.1
PATH_BUILD=build/
FILE_COMMAND=terragrunt-atlantis-config
FILE_ARCH=darwin_amd64
S3_BUCKET_NAME=cloudfront-origin-homebrew-tap-transcend-io
PROFILE=transcend-prod

# Determine the arch/os combos we're building for
XC_ARCH=amd64 arm
XC_OS=linux darwin windows

.PHONY: clean
clean:
	rm -rf ./build
	rm -rf '$(HOME)/bin/$(FILE_COMMAND)'

.PHONY: build
build: clean
	@$(GOPATH)/bin/goxc \
    -bc="darwin,amd64" \
    -pv=$(VERSION) \
    -d=$(PATH_BUILD) \
    -build-ldflags "-X main.VERSION=$(VERSION)"

.PHONY: build-all
build-all: clean
	@$(GOPATH)/bin/goxc \
	-os="$(XC_OS)" \
	-arch="$(XC_ARCH)" \
    -pv=$(VERSION) \
    -d=$(PATH_BUILD) \
    -build-ldflags "-X main.VERSION=$(VERSION)"

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
sign:  build-all
	rm -f $(PATH_BUILD)${VERSION}/SHA256SUMS
	shasum -a256 $(PATH_BUILD)${VERSION}/* > $(PATH_BUILD)${VERSION}/SHA256SUMS

.PHONY: install
install:
	install -d -m 755 '$(HOME)/bin/'
	install $(PATH_BUILD)$(FILE_COMMAND)/$(VERSION)/$(FILE_COMMAND)_$(VERSION)_$(FILE_ARCH) '$(HOME)/bin/$(FILE_COMMAND)'