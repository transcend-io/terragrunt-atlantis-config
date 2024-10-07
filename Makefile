VERSION=1.19.0
PATH_BUILD=build/
FILE_COMMAND=terragrunt-atlantis-config
FILE_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

# Determine the arch/os combos we're building for
XC_ARCH=amd64 arm64
XC_OS=linux darwin windows

.PHONY: clean
clean:
	rm -rf ./build
	rm -rf "$(HOME)/.local/bin/$(FILE_COMMAND)"

.PHONY: build
build: clean
	CGO_ENABLED=0 \
	go build \
	-trimpath \
	-mod=readonly \
	-modcacherw \
	-ldflags "-X main.VERSION=$(VERSION)" \
	-o $(PATH_BUILD)$(VERSION)/$(FILE_COMMAND)_$(VERSION)_$(FILE_ARCH)

.PHONY: build-all
build-all: clean
	for arch in $(XC_ARCH); do \
		for os in $(XC_OS); do \
			echo "Building for '$$os/$$arch'" ; \
			ext="" ; [ "$$os" = "windows" ] && ext=".exe" ; \
			CGO_ENABLED=0 \
			GOARCH=$$arch \
			GOOS=$$os \
			go build \
			-trimpath \
			-mod=readonly \
			-modcacherw \
			-ldflags "-X main.VERSION=$(VERSION)" \
			-o $(PATH_BUILD)$(VERSION)/$(FILE_COMMAND)_$(VERSION)_$${os}_$${arch}$${ext} ; \
		done \
	done

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
	rm -f $(PATH_BUILD)$(VERSION)/SHA256SUMS
	shasum -a256 $(PATH_BUILD)$(VERSION)/* > $(PATH_BUILD)$(VERSION)/SHA256SUMS

.PHONY: install
install:
	install -d -m 755 '$(HOME)/.local/bin/'
	install $(PATH_BUILD)$(VERSION)/$(FILE_COMMAND)_$(VERSION)_$(FILE_ARCH) '$(HOME)/.local/bin/$(FILE_COMMAND)'
