package tests_test

import (
	"flag"
	"fmt"
	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	ktests "kubevirt.io/kubevirt/tests"
	"time"
)

//// The list of supported Operating Systems
//// name : { registryDisk, cloudinit }
var osDict = map[string][]string{
	"testvm-fedora": {"kubevirt/fedora-cloud-registry-disk-demo:latest", "#cloud-config\npassword: fedora\nchpasswd: { expire: False }"},
}

var clients = [2]string{"oc", "kubectl"}

const (
	templatePath  = "tests/manifests/vm-fedora-template.yml"
	temporaryJson = "/tmp/tmp-vm.yml"
	memory        = "1024Mi"
	cpuCores      = "1"

	ipQuery    = "-o=jsonpath='{.status.interfaces[0].ipAddress}'"
	phaseQuery = "-o=jsonpath='{.status.phase}'"
)

var _ = Describe("Regression and Functional tests of VMs and VMIs", func() {

	flag.Parse()

	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	Describe("Creating VMs with different OS from templates via oc and kubectl", func() {

		// Run tests for each OS:
		for k := range osDict {
			registryDisk := osDict[k][0]
			cloudInit := osDict[k][1]

			var vm tests.VirtualMachine
			vm.Name = k
			vm.Namespace = ktests.NamespaceTestDefault

			// Run for different clients: oc and kubectl
			for _, cl := range clients {

				client := cl

				Context("Create/Delete "+vm.Name+" VM from template via "+client+" command", func() {

					It("Verify simple functionality of "+vm.Name+" VM: create, start, check status, IP, stop and delete.", func() {
						By("Creating VM via " + client)
						tests.ProcessTemplateWithParameters(templatePath, temporaryJson, "NAME="+vm.Name, "CPU_CORES="+cpuCores, "MEMORY="+memory, "IMAGE_NAME="+registryDisk, "CLOUD_INIT="+cloudInit)
						_, _, err := ktests.RunCommand(client, "create", "-f", temporaryJson)
						Expect(err).ToNot(HaveOccurred(), "VM 'creating' command should be executed without errors")

						By("Starting VM")
						_, _, err = vm.Start()
						Expect(err).ToNot(HaveOccurred(), "VM should start without errors")

						By("Waiting until VMI is running")
						tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vm.Name, phaseQuery, "Running")

						ipAddress := tests.GetVirtualMachineSpecificParameters("vmi", vm.Name, ipQuery)
						// removing quotes
						ipAddress = ipAddress[1 : len(ipAddress)-1]

						By("Waiting for the console and check IP address")
						switch vm.Name {
						case "testvm-fedora":
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

						By("Checking if the VM's VNC server gives the valid response")
						response, err := tests.VNCConnection(vm.Namespace, vm.Name)
						Expect(err).ToNot(HaveOccurred(), "Should open VNC connection to VMI %q in %s namespace", vm.Name, vm.Namespace)
						Expect(response).To(Equal("RFB 003.008"), "Should receive valid response from the VNC connection to the VMI %q in %s namespace", vm.Name, vm.Namespace)

						By("Stopping VM")
						_, _, err = vm.Stop()
						Expect(err).ToNot(HaveOccurred(), "VM should stopp be executed without errors")

						By("Deleting VM via " + client)
						_, _, err = ktests.RunCommand(client, "delete", "-f", temporaryJson)
						Expect(err).ToNot(HaveOccurred(), "VM 'deleting' command should be executed without errors")

						By("Verifying that VM was removed")
						tests.WaitUntilResourceDeleted("vm", vm.Name)
					})
				})
			}
		}
	})

	Describe("VM Lifecycle - negative tests", func() {

		Context("Delete non existing VM", func() {
			It("Try to delete non existing VM", func() {
				By("Deleting VM and should get an error")
				_, _, err := ktests.RunCommand("oc", "delete", "vm", "TESTVM-delete")
				Expect(err).To(HaveOccurred(), "Should get an error, as VM does not exists")
			})
		})

		Context("Create VM with existing name", func() {
			It("Try to create VM with existing name", func() {
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
})
