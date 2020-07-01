VERSION=0.4.1
PATH_BUILD=build/
FILE_COMMAND=terragrunt-atlantis-config
FILE_ARCH=darwin_amd64
S3_BUCKET_NAME=cloudfront-origin-homebrew-tap-transcend-io
PROFILE=transcend-prod

clean:
	rm -rf ./build
	rm -rf '$(HOME)/bin/$(FILE_COMMAND)'

build: clean
	@$(GOPATH)/bin/goxc \
    -bc="darwin,amd64" \
    -pv=$(VERSION) \
    -d=$(PATH_BUILD) \
    -build-ldflags "-X main.VERSION=$(VERSION)"

test:
	go test -v ./...

version:
	@echo $(VERSION)

shasum:  build
	shasum -a256 $(PATH_BUILD)$(VERSION)/$(FILE_COMMAND)_$(VERSION)_$(FILE_ARCH).zip

install:
	install -d -m 755 '$(HOME)/bin/'
	install $(PATH_BUILD)$(VERSION)/$(FILE_ARCH)/$(FILE_COMMAND) '$(HOME)/bin/$(FILE_COMMAND)'

publish: build
	AWS_PROFILE=$(PROFILE) aws s3 sync $(PATH_BUILD)/$(VERSION) s3://$(S3_BUCKET_NAME)/$(FILE_COMMAND)/$(VERSION)

ci_publish: build
	aws s3 sync $(PATH_BUILD)$(VERSION) s3://$(S3_BUCKET_NAME)/$(FILE_COMMAND)/$(VERSION)