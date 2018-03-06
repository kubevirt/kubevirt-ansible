ROOT_PATH        ?= "$(shell pwd)"
ORG              ?= ansibleplaybookbundle
TAG              ?= latest
REGISTRY         ?= docker.io

kubevirt-apb:
	$(eval IMAGE_NAME?=kubevirt-apb)
	pushd $(ROOT_PATH); docker build --tag $(REGISTRY)/$(ORG)/$(IMAGE_NAME):$(TAG) -f ${ROOT_PATH}/kubevirt-apb/Dockerfile $(ROOT_PATH); popd

.PHONY: kubevirt-apb
