package tests_test

import (
	"flag"
	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	ktests "kubevirt.io/kubevirt/tests"
	"time"
)

const (
	templatePath  = "tests/manifests/vm-fedora-template.yml"
	temporaryJson = "/tmp/tmp-vm.yml"

	ipQuery    = "-o=jsonpath='{.status.interfaces[0].ipAddress}'"
	phaseQuery = "-o=jsonpath='{.status.phase}'"
)

var _ = Describe("[rfe_id:273][crit:medium][vendor:cnv-qe@redhat.com][level:component]Create/Delete VM from Templates and yaml files", func() {

	flag.Parse()

	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	var vm tests.VirtualMachine
	vm.Name = "vm-fedora"
	vm.Namespace = ktests.NamespaceTestDefault

	Describe("Create and Delete VM from Templates", func() {
		table.DescribeTable("Create and Delete via oc or kubectl", func(client string) {
			By("Creating VM via " + client)
			tests.ProcessTemplateWithParameters(templatePath, temporaryJson, "NAME="+vm.Name)
			_, _, err := ktests.RunCommand(client, "create", "-f", temporaryJson)
			Expect(err).ToNot(HaveOccurred(), "VM 'creating' command should be executed without errors")

			By("Starting VM")
			_, _, err = vm.Start()
			Expect(err).ToNot(HaveOccurred(), "VM should start without errors")

			By("Waiting until VMI is running")
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vm.Name, phaseQuery, "Running")

			By("Delete VM")
			_, _, err = ktests.RunCommand(client, "delete", "vm", vm.Name)
			Expect(err).ToNot(HaveOccurred(), "VM 'deleting' command should be executed without errors")

			By("Verifying that POD was removed")
			tests.WaitUntilResourceDeleted("pod", vm.Name)

			By("Verifying that VM was removed")
			tests.WaitUntilResourceDeleted("vm", vm.Name)
		},
			table.Entry("[test_id:502]using oc", "oc"),
			table.Entry("[test_id:264]using kubectl", "kubectl"),
		)
	})

	Describe("Create and Delete VM via yaml file, verify console and vnc", func() {
		It("[test_id:241]should create and run VM, verify VNC and console connection, IP address", func() {
			By("Create and run VM")
			_, _, err := ktests.RunCommand("oc", "create", "-f", temporaryJson)
			Expect(err).ToNot(HaveOccurred(), "VM 'creating' command should be executed without errors")
			_, _, err = vm.Start()
			Expect(err).ToNot(HaveOccurred(), "VM should start without errors")
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vm.Name, phaseQuery, "Running")

			By("Get IP address from VMI describe")
			ipAddress := tests.GetVirtualMachineSpecificParameters("vmi", vm.Name, ipQuery)
			ipAddress = ipAddress[1 : len(ipAddress)-1]

			By("Verify the console and IP on Fedora")
			expecter, err := tests.LoggedInFedoraExpecter(vm.Name, tests.NamespaceTestDefault, 360)
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()
			_, err = expecter.ExpectBatch([]expect.Batcher{
				&expect.BSnd{S: "ifconfig\n"},
				&expect.BExp{R: ipAddress},
			}, 200*time.Second)
			Expect(err).ToNot(HaveOccurred(), "IP should be the same as in POD description")

			By("Verify VNC connection")
			response, err := tests.VNCConnection(vm.Namespace, vm.Name)
			Expect(err).ToNot(HaveOccurred(), "Should open VNC connection to VMI %q in %s namespace", vm.Name, vm.Namespace)
			Expect(response).To(Equal("RFB 003.008"), "Should receive valid response from the VNC connection to the VMI %q in %s namespace", vm.Name, vm.Namespace)

			By("Delete VM via yaml file")
			_, _, err = ktests.RunCommand("oc", "delete", "-f", temporaryJson)
			Expect(err).ToNot(HaveOccurred(), "VM 'deleting' command should be executed without errors")

			By("Verify the POD and VM were deleted")
			tests.WaitUntilResourceDeleted("pod", vm.Name)
			tests.WaitUntilResourceDeleted("vm", vm.Name)
		})
	})

	Describe("VM Lifecycle - negative tests", func() {
		It("[test_id:233][posneg:negative]Delete non existing VM", func() {
			_, _, err := ktests.RunCommand("oc", "delete", "vm", "TESTVM-delete")
			Expect(err).To(HaveOccurred(), "Should get an error, as VM does not exists")
		})

		It("[test_id:243][posneg:negative]Create VM with existing name", func() {
			By("Creating first VM")
			tests.ProcessTemplateWithParameters(templatePath, temporaryJson, "NAME=new-vm")
			_, _, err := ktests.RunCommand("oc", "create", "-f", temporaryJson)
			Expect(err).ToNot(HaveOccurred(), "VM 'creating' command should be executed without errors")

			By("Creating second VM with the same name")
			_, _, err = ktests.RunCommand("oc", "create", "-f", temporaryJson)
			Expect(err).To(HaveOccurred(), "Should get an error, as VM with the same name already exists")
		})
	})
})
