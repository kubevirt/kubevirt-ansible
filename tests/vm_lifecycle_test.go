package tests_test

import (
	"flag"
	"fmt"
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

var _ = FDescribe("[rfe_id:273][crit:medium][vendor:cnv-qe@redhat.com][level:component]Create/Import VM from Template", func() {

	flag.Parse()

	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	Describe("Creating VMs with different OS from templates via oc and kubectl", func() {

		var vm tests.VirtualMachine
		vm.Name = "vm-fedora"
		vm.Namespace = ktests.NamespaceTestDefault

		table.DescribeTable("Create/Delete "+vm.Name+" VM from template", func(client string) {
			Context("should create and run VM", func() {
				By("Creating VM via " + client)
				tests.ProcessTemplateWithParameters(templatePath, temporaryJson, "NAME="+vm.Name)
				_, _, err := ktests.RunCommand(client, "create", "-f", temporaryJson)
				Expect(err).ToNot(HaveOccurred(), "VM 'creating' command should be executed without errors")

				By("Starting VM")
				_, _, err = vm.Start()
				Expect(err).ToNot(HaveOccurred(), "VM should start without errors")

				By("Waiting until VMI is running")
				tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vm.Name, phaseQuery, "Running")
			})

			Context("[test_id:241]should verify that VM receives pod IP address", func() {
				By("Get IP address from VMI describe")
				ipAddress := tests.GetVirtualMachineSpecificParameters("vmi", vm.Name, ipQuery)
				ipAddress = ipAddress[1 : len(ipAddress)-1]

				By("Waiting for the console and check IP address")
				switch vm.Name {
				case "vm-fedora":
					By("Verifying the console and IP on Fedora")
					expecter, err := tests.LoggedInFedoraExpecter(vm.Name, tests.NamespaceTestDefault, 360)
					Expect(err).ToNot(HaveOccurred())
					defer expecter.Close()
					_, err = expecter.ExpectBatch([]expect.Batcher{
						&expect.BSnd{S: "ifconfig\n"},
						&expect.BExp{R: ipAddress},
					}, 200*time.Second)
					Expect(err).ToNot(HaveOccurred(), "IP should be the same as in POD description")
				default:
					fmt.Println("Can't verify console on this VM")
				}
			})

			Context("[test_id:242]should verify VNC connetion to running VM", func() {
				response, err := tests.VNCConnection(vm.Namespace, vm.Name)
				Expect(err).ToNot(HaveOccurred(), "Should open VNC connection to VMI %q in %s namespace", vm.Name, vm.Namespace)
				Expect(response).To(Equal("RFB 003.008"), "Should receive valid response from the VNC connection to the VMI %q in %s namespace", vm.Name, vm.Namespace)
			})

			if client == "oc" {
				Context("[test_id:300]Delete VM via oc command", func() {
					_, _, err := ktests.RunCommand("oc", "delete", "vm", vm.Name)
					Expect(err).ToNot(HaveOccurred(), "VM 'deleting' command should be executed without errors")
				})
			} else {
				Context("[test_id:263]Delete VM via oc command using yaml file", func() {
					_, _, err := ktests.RunCommand("oc", "delete", "-f", temporaryJson)
					Expect(err).ToNot(HaveOccurred(), "VM 'deleting' command should be executed without errors")
				})
			}

			Context("verify that all resources were removed", func() {
				By("Verifying that POD was removed")
				tests.WaitUntilResourceDeleted("pod", vm.Name)

				By("Verifying that VM was removed")
				tests.WaitUntilResourceDeleted("vm", vm.Name)
			})
		},
			table.Entry("[test_id:502]using oc", "oc"),
			table.Entry("[test_id:264]using kubectl", "kubectl"),
		)
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
