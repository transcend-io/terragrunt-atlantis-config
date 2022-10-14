# Will use the default value beta if the environment variable doesn't exist
TAG_VERSION?=beta

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
	CGO_ENABLED=0 \
	goxc \
    -bc="darwin,amd64" \
    -pv=$(TAG_VERSION) \
    -d=$(PATH_BUILD) \
    -build-ldflags "-X main.VERSION=$(TAG_VERSION)"

.PHONY: build-all
build-all: clean
	CGO_ENABLED=0 \
	goxc \
	-os="$(XC_OS)" \
	-arch="$(XC_ARCH)" \
    -pv=$(TAG_VERSION) \
    -d=$(PATH_BUILD) \
    -build-ldflags "-X main.VERSION=$(TAG_VERSION)"

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
	@echo $(TAG_VERSION)

.PHONY: sign
sign:  build-all
	rm -f $(PATH_BUILD)${TAG_VERSION}/SHA256SUMS
	cd $(PATH_BUILD)${TAG_VERSION} && shasum -a256 * > SHA256SUMS

.PHONY: install
install:
	install -d -m 755 '$(HOME)/bin/'
	install $(PATH_BUILD)$(FILE_COMMAND)/$(TAG_VERSION)/$(FILE_COMMAND)_$(TAG_VERSION)_$(FILE_ARCH) '$(HOME)/bin/$(FILE_COMMAND)'
