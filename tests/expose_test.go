package tests_test


import (
	"flag"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktests "kubevirt.io/kubevirt/tests"
	"github.com/davecgh/go-spew/spew"
	"kubevirt.io/kubevirt-ansible/tests"
	"kubevirt.io/kubevirt/pkg/kubecli"
)

const vmName = "vm-cirros"

var _ = Describe("D1", func() {

	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	ktests.SkipIfNoOc()

	Context("C1", func() {

		It("the VM should be created successfully", func() {
			tests.CreateResourceWithFilePathTestNamespace("tests/manifests/vm-cirros.yml")
		})

		It("the VM should be started by virtctl", func() {
			args := []string{"start", vmName, "-n", tests.NamespaceTestDefault}
			_, err := ktests.RunCommand("virtctl", args...)
			Expect(err).ToNot(HaveOccurred())
		})

		Specify("the VMI should be running", func() {
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vmName, "-o=jsonpath='{.status.phase}'", "Running")
			args := []string{"get", "vmi", vmName, "-o", "yaml", "-n", tests.NamespaceTestDefault}
			out, err := ktests.RunCommand("oc", args...)
			spew.Dump("====================")
			Expect(err).ToNot(HaveOccurred())
			spew.Dump(out)
		})

		It("the VM should be exposed by virtctl via ClusterIP", func() {
			args := []string{"expose", "vmi", vmName, "-n", tests.NamespaceTestDefault, "--name=myvm-ssh", "--port=22", "--type=ClusterIP"}
			_, err := ktests.RunCommand("virtctl", args...)
			Expect(err).ToNot(HaveOccurred())
		})

		It("the VMI should be able to login by ssh", func() {
			By("Create an testpod as an ssh client to connect to VMI")
			args := []string{"run", "testpod", "-n", tests.NamespaceTestDefault, "--image=shiywang/test-ssh", "--", "sleep", "3600"}
			_, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			tests.WaitUntilResourceReadyByLabelTestNamespace("pod", "deploymentconfig=testpod", "-o=jsonpath='{.items[*].status.phase}'", "Running")

			By("Fetch the name of testpod")
			args = []string{"get", "pod", "-l", "deploymentconfig=testpod", "-n", tests.NamespaceTestDefault, "-o=jsonpath='{.items[*].metadata.name}'"}
			podName, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			//oc get svc myvm-ssh -o=jsonpath='{.spec.clusterIP}'
			By("Fetch the clusterIP of service myvm-ssh")
			args = []string{"get", "svc", "myvm-ssh", "-o=jsonpath='{.spec.clusterIP}'"}
			clusterIP, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())


			spew.Dump(podName[1:len(podName)-1])
			By("Run exec ssh command inside the testpod")
			args = []string{"exec", podName[1:len(podName)-1], "-n", tests.NamespaceTestDefault, "-i", "-t", "--", "sshlogin", clusterIP[1:len(clusterIP)-1], "cirros", "gocubsgo"}
			out, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			spew.Dump(out)
			Expect(out).Should(ContainSubstring("vm-cirros"))
		})


		It("the VMI should be able to login by virtctl console", func() {
			vmi, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Get(vmName, &metav1.GetOptions{})
			expecter, err := ktests.LoggedInCirrosExpecter(vmi)
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()
		})

		It("the VM should be stop by virtctl", func() {
			args := []string{"stop", vmName, "-n", tests.NamespaceTestDefault}
			_, err := ktests.RunCommand("virtctl", args...)
			Expect(err).ToNot(HaveOccurred())
		})

	})
})
