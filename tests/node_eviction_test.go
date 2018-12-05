package tests_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	ktests "kubevirt.io/kubevirt/tests"
)

var _ = Describe("Node Eviction", func() {
	var vmi tests.VirtualMachine
	vmi.Namespace = ktests.NamespaceTestDefault

	BeforeEach(func() {
		ktests.BeforeTestCleanup()

		args := []string{"get", "node", "--selector=node-role.kubernetes.io/compute=true", "--output=jsonpath={.items..metadata.name}"}
		output, _, err := ktests.RunCommand("oc", args...)
		Expect(err).ToNot(HaveOccurred())

		var count int
		nodes := strings.Fields(output)
		for _, node := range nodes {
			args := []string{"get", "node", node}
			output, _, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			if !strings.Contains(output, "SchedulingDisabled") && !strings.Contains(output, "NotReady") {
				count++
			}
		}
		if count < 2 {
			Skip("not enough schedulable nodes for this testing")
		}

	})

	AfterEach(func() {
		// nodes can be in drained status after some tests is failed, ensure uncordoning them after each tests.
		uncordonNodes()
	})

	Describe("VirtualMachineInstance Evictions", func() {
		BeforeEach(func() {
			vmi.Manifest = "tests/manifests/vmi-ephemeral.yaml"
			vmi.Name = "vmi-ephemeral"
			vmi.Type = "vmi"
		})
		It("virtualmachine instance will not be re-scheduled", func() {
			By("Create the virtualmachine instance")
			_, err := vmi.Create()
			Expect(err).ToNot(HaveOccurred())

			By("Check that the virtualmachine instance is running")
			Eventually(func() bool {
				output, err := vmi.IsRunning()
				Expect(err).ToNot(HaveOccurred())
				return output
			}, time.Minute*10).Should(BeTrue())

			By("Retrieve the vmi's node")
			nodeName, err := vmi.GetVMInfo("{{.status.nodeName}}")
			Expect(err).ToNot(HaveOccurred())

			By("Drain the vmi's node")
			args := []string{"adm", "drain", nodeName, "--delete-local-data", "--ignore-daemonsets", "--force", "--pod-selector=kubevirt.io=virt-launcher"}
			_, _, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the virtual machine instance is Failed")
			output, err := vmi.GetVMInfo("{{.status.phase}}")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal("Failed"))

			By("Verify that the vmi's node is unschedulable")
			args = []string{"get", "node", nodeName, "--template", "{{.spec.unschedulable}}"}
			output, _, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal("true"))

			By("Uncordon the vmi's node")
			args = []string{"adm", "uncordon", nodeName}
			_, _, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the node is schedulable again")
			args = []string{"get", "node", nodeName}
			output, _, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(ContainSubstring("SchedulingDisabled"))
		})
	})

	Describe("VirtualMachineInstanceReplicaSet Eviction", func() {
		BeforeEach(func() {
			vmi.Name = "vmi-replicaset-cirros"
			vmi.Manifest = "tests/manifests/vmi-replicaset-cirros.yaml"
			vmi.Type = "vmi"
		})
		It("virtualmachineinstance owned by a replicaset will be re-scheduled", func() {
			By("Create the virtualmachine instance replicaset")
			_, err := vmi.Create()
			Expect(err).ToNot(HaveOccurred())

			By("Check that the virtualmachine instance is running")
			args := []string{"get", vmi.Type, "--selector=kubevirt.io/vmReplicaSet=vmi-replicaset-cirros", "--output=jsonpath={.items..metadata.name}"}
			vmis, _, err := ktests.RunCommandWithNS(vmi.Namespace, "oc", args...)
			Expect(err).ToNot(HaveOccurred())

			vmiList := strings.Fields(vmis)
			for _, vminame := range vmiList {
				vmi.Name = vminame
				Eventually(func() bool {
					output, err := vmi.IsRunning()
					Expect(err).ToNot(HaveOccurred())
					return output
				}, time.Minute*10).Should(BeTrue())
			}

			By("Retrieve the vmi's node")
			nodeName, err := vmi.GetVMInfo("{{.status.nodeName}}")
			Expect(err).ToNot(HaveOccurred())

			By("Drain the vmi's node")
			args = []string{"adm", "drain", nodeName, "--delete-local-data", "--ignore-daemonsets=true", "--force", "--pod-selector=kubevirt.io=virt-launcher"}
			_, _, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the vmi's node is unschedulable")
			args = []string{"get", "node", nodeName, "--template", "{{.spec.unschedulable}}"}
			Eventually(func() string {
				output, _, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return output
			}, time.Minute*2).Should(Equal("true"))

			By("Verify that the vmi is running on other node")
			args = []string{"get", vmi.Type, "--selector=kubevirt.io/vmReplicaSet=vmi-replicaset-cirros", "--output=jsonpath={.items..metadata.name}"}
			vmis, _, err = ktests.RunCommandWithNS(vmi.Namespace, "oc", args...)
			Expect(err).ToNot(HaveOccurred())

			vmiList = strings.Fields(vmis)
			for _, vmiName := range vmiList {
				vmi.Name = vmiName
				Eventually(func() bool {
					output, err := vmi.IsRunning()
					Expect(err).ToNot(HaveOccurred())
					return output
				}, time.Minute*2).Should(BeTrue())

				output, err := vmi.GetVMInfo("{{.status.nodeName}}")
				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(Equal(nodeName))
			}

			By("Uncordon the vmi's node")
			args = []string{"adm", "uncordon", nodeName}
			_, _, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the node is schedulable again")
			args = []string{"get", "node", nodeName}
			output, _, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(ContainSubstring("SchedulingDisabled"))
		})
	})

	Describe("VirtualMachine Eviction", func() {
		BeforeEach(func() {
			vmi.Manifest = "tests/manifests/vm-cirros.yaml"
			vmi.Name = "vm-cirros"
			vmi.Type = "vmi"
		})
		It("virtualmachineinstance owned by a virtual machine will be re-scheduled", func() {
			By("Create the virtualmachine instance's vm")
			_, err := vmi.Create()
			Expect(err).ToNot(HaveOccurred())

			By("Start the virtual machine via virtctl")
			_, err = vmi.Start()
			Expect(err).ToNot(HaveOccurred())

			By("Check that the virtualmachine instance is running")
			Eventually(func() bool {
				output, err := vmi.IsRunning()
				Expect(err).ToNot(HaveOccurred())
				return output
			}, time.Minute*10).Should(BeTrue())

			By("Retrieve the vmi's node")
			nodeName, err := vmi.GetVMInfo("{{.status.nodeName}}")
			Expect(err).ToNot(HaveOccurred())

			By("Drain the vmi's node")
			args := []string{"adm", "drain", nodeName, "--delete-local-data", "--ignore-daemonsets=true", "--force", "--pod-selector=kubevirt.io=virt-launcher"}
			_, _, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the vmi's node is unschedulable")
			args = []string{"get", "node", nodeName, "--template", "{{.spec.unschedulable}}"}
			Eventually(func() (string, string, error) {
				return ktests.RunCommand("oc", args...)
			}, time.Minute*2).Should(Equal("true"))

			By("Verify that the vmi is running on other node")
			Eventually(func() bool {
				output, err := vmi.IsRunning()
				Expect(err).ToNot(HaveOccurred())
				return output
			}, time.Minute*2).Should(BeTrue())

			By("Uncordon the vmi's node")
			args = []string{"adm", "uncordon", nodeName}
			_, _, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the node is schedulable again")
			args = []string{"get", "node", nodeName}
			output, _, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(ContainSubstring("SchedulingDisabled"))
		})
	})
})

func uncordonNodes() {
	args := []string{"get", "node", "--selector=node-role.kubernetes.io/compute=true", "--output=jsonpath={.items..metadata.name}"}
	output, _, err := ktests.RunCommand("oc", args...)
	Expect(err).ToNot(HaveOccurred())

	nodes := strings.Fields(output)
	for _, node := range nodes {
		args := []string{"get", "node", node}
		output, _, err := ktests.RunCommand("oc", args...)
		Expect(err).ToNot(HaveOccurred())

		if strings.Contains(output, "SchedulingDisabled") {
			args = []string{"adm", "uncordon", node}
			_, _, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
		}
	}
}
