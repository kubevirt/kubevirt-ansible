module kubevirt.io/kubevirt-ansible

require (
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20191011121108-aa519ddbe484 // indirect
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/go-openapi/spec v0.19.4 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.1.1 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/golang/mock v1.3.1 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/goexpect v0.0.0-20191001010744-5b6988669ffa
	github.com/google/goterm v0.0.0-20190703233501-fc88cf888a3f // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v0.0.0-20191119172530-79f836b90111 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-runewidth v0.0.6 // indirect
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v3.9.0+incompatible // indirect
	github.com/openshift/custom-resource-status v0.0.0-20190822192428-e62f2f3b79f3 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/spf13/cobra v0.0.5 // indirect
	golang.org/x/crypto v0.0.0-20191128160524-b544559bb6d1
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3 // indirect
	golang.org/x/net v0.0.0-20191126235420-ef20fe5d7933 // indirect
	golang.org/x/sys v0.0.0-20191128015809-6d18c012aee9 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/grpc v1.25.1 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
	k8s.io/api v0.0.0-20191121015604-11707872ac1c
	k8s.io/apiextensions-apiserver v0.0.0-20191121021419-88daf26ec3b8 // indirect
	k8s.io/apimachinery v0.0.0-20191123233150-4c4803ed55e3
	k8s.io/client-go v11.0.0+incompatible // indirect
	kubevirt.io/containerized-data-importer v1.11.0 // indirect
	kubevirt.io/kubevirt v0.13.7
	kubevirt.io/qe-tools v0.1.2
)

replace (
	github.com/go-kit/kit => github.com/go-kit/kit v0.3.0
	github.com/k8snetworkplumbingwg/network-attachment-definition-client => github.com/booxter/network-attachment-definition-client v0.0.0-20181123022110-379ff533bf29
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.4.1-0.20170829012221-11459a886d9c
	github.com/onsi/gomega => github.com/onsi/gomega v1.2.1-0.20170829124025-dcabb60a477c
	k8s.io/api => k8s.io/api v0.0.0-20180712090710-2d6f90ab1293
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20180808065829-408db4a50408
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20180621070125-103fd098999d
	k8s.io/client-go => k8s.io/client-go v0.0.0-20180806134042-1f13a808da65
	kubevirt.io/kubevirt => kubevirt.io/kubevirt v0.13.7
	kubevirt.io/qe-tools => kubevirt.io/qe-tools v0.1.2
)
