deps-update:
	glide cc && glide update --strip-vendor
	hack/dep-prune.sh

test:
	hack/dockerized.sh "hack/build-tests.sh"
	hack/run-tests.sh

.PHONY: deps-update test
