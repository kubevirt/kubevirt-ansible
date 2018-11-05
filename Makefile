deps-update:
	SYNC_VENDOR=true hack/dockerized.sh "glide cc && glide update --strip-vendor"
	hack/dep-prune.sh

distclean:
	hack/dockerized.sh "rm -rf vendor/ && rm -f .glide.*.hash && glide cc"
	rm -rf vendor/

generate:
	SYNC_GENERATED=true hack/dockerized.sh "hack/generate.sh"

check:
	hack/dockerized.sh "hack/check.sh"

build-tests:
	hack/dockerized.sh "hack/build-tests.sh"

test:
	hack/run-tests.sh

.PHONY: deps-update distclean generate check build-tests test
