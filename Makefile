deps-update:
	glide cc && glide update --strip-vendor
	hack/dep-prune.sh

test:
	cd tests/ && ../hack/functest.sh

.PHONY: deps-update test
