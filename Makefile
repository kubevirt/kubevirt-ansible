deps-update:
	SYNC_VENDOR=true hack/dockerized.sh "glide cc && glide update --strip-vendor"
	hack/dep-prune.sh

distclean:
	hack/dockerized.sh "rm -rf vendor/ && rm -f .glide.*.hash && glide cc"
	rm -rf vendor/

generate-tests:
	hack/dockerized.sh "hack/build-tests.sh"

test:
	hack/run-tests.sh

.PHONY: deps-update distclean generate-tests test
