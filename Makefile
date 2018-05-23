deps-update:
	glide cc && glide update --strip-vendor
	hack/dep-prune.sh

generate-tests:
	hack/dockerized.sh "hack/build-tests.sh"

test:
	hack/run-tests.sh

.PHONY: deps-update generate-tests test
