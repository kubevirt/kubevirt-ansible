deps-update:
	glide cc && glide update --strip-vendor
	hack/dep-prune.sh

test:
	cd tests/ && go test -v ./...

.PHONY: deps-update test
