package tests_test

import (
	"flag"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	expect "github.com/google/goexpect"

	k8sv1 "k8s.io/api/core/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	v1 "kubevirt.io/kubevirt/pkg/api/v1"
	ktests "kubevirt.io/kubevirt/tests"

	"k8s.io/apimachinery/pkg/api/resource"
	"kubevirt.io/kubevirt/pkg/kubecli"
)

const (
	denyallYaml   = "tests/manifests/networkpolicy-deny-all.yml"
	allowhttpYaml = "tests/manifests/networkpolicy-allow-http.yml"
)

var _ = Describe("[rfe_id:150][crit:high][vendor:cnv-qe@redhat.com][level:component]Networkpolicy", func() {
	flag.Parse()

	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	var vmia *v1.VirtualMachineInstance
	var vmib *v1.VirtualMachineInstance

	ktests.BeforeAll(func() {
		// Flannel networking model doesn't support network policy
		ktests.SkipIfUseFlannel(virtClient)
		ktests.SkipIfNotUseNetworkPolicy(virtClient)
		ktests.BeforeTestCleanup()
	})

	Context("Networkpolicy allow http port", func() {
		It("[test_id:CNV-369] Set connectivity to only allow from http port", func() {

			curlReq := func(ip string, port string, vmi *v1.VirtualMachineInstance, resp string) {
				ktests.WaitUntilVMIReady(vmi, ktests.LoggedInCirrosExpecter)
				err := ktests.CheckForTextExpecter(vmi, []expect.Batcher{
					&expect.BSnd{S: fmt.Sprintf("curl --silent --connect-timeout 5 %s%s  | grep hellokubevirt | wc -l \n", ip, port)},
					&expect.BExp{R: resp},
				}, 60)
				Expect(err).ToNot(HaveOccurred())
			}
			By("Create VMIs")
			userData := "#cloud-config\npassword: fedora\nchpasswd: { expire: False }\n"
			vmia = ktests.NewRandomVMIWithEphemeralDiskAndUserdata(ktests.ContainerDiskFor(ktests.ContainerDiskFedora), userData)
			ktests.AddExplicitPodNetworkInterface(vmia)

			vmia.Spec.Domain.Resources.Requests[k8sv1.ResourceName("memory")] = resource.MustParse("1024M")
			_, err = virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Create(vmia)
			Expect(err).ToNot(HaveOccurred())

			vmib = ktests.NewRandomVMIWithEphemeralDiskAndUserdata(ktests.ContainerDiskFor(ktests.ContainerDiskCirros), "#!/bin/bash\necho 'hello'\n")
			_, err = virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Create(vmib)
			Expect(err).ToNot(HaveOccurred())

			ktests.WaitForSuccessfulVMIStart(vmia)
			vmia, err = virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Get(vmia.Name, &v13.GetOptions{})
			Expect(err).ToNot(HaveOccurred())

			By("Install httpd and start the service inside the VMI")
			ip := vmia.Status.Interfaces[0].IP

			ktests.WaitUntilVMIReady(vmia, ktests.LoggedInFedoraExpecter)
			err = ktests.CheckForTextExpecter(vmia, []expect.Batcher{
				&expect.BSnd{S: "\n"},
				&expect.BExp{R: "#"},
				&expect.BSnd{S: "yum install httpd -y\n"},
				&expect.BExp{R: "#"},
				&expect.BSnd{S: "echo hellokubevirt | sudo tee /var/www/html/index.html\n"},
				&expect.BExp{R: "hellokubevirt"},
				&expect.BSnd{S: "sed -i 's/Listen 80/Listen 80\nListen 81/g' /etc/httpd/conf/httpd.conf\n"},
				&expect.BExp{R: "#"},
				&expect.BSnd{S: "systemctl start httpd\n"},
				&expect.BExp{R: "#"},
			}, 180)
			Expect(err).ToNot(HaveOccurred())

			By("Validate both 80 and 81 ports are available")
			httpport := ":80"
			otherport := ":81"
			resp := "1"
			curlReq(ip, httpport, vmib, resp)
			curlReq(ip, otherport, vmib, resp)

			By("Create deny-all networkpolicy and validate all the traffic are limited")
			tests.CreateResourceWithFilePath(denyallYaml)
			resp = "0"
			curlReq(ip, httpport, vmib, resp)
			curlReq(ip, otherport, vmib, resp)

			By("Create allow-tcp networkpolicy and validate only http port is available")
			tests.CreateResourceWithFilePath(allowhttpYaml)
			curlReq(ip, otherport, vmib, resp)
			resp = "1"
			curlReq(ip, httpport, vmib, resp)
		})
	})
})
