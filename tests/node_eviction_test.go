package tests_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"kubevirt.io/kubevirt-ansible/tests"
	ktests "kubevirt.io/kubevirt/tests"
)

var _ = Describe("Node Eviction", func() {
	ns := ktests.NamespaceTestDefault

	var vmi tests.VMManifest

	BeforeEach(func() {
		ktests.BeforeTestCleanup()

		args := []string{"get", "node", "--selector=node-role.kubernetes.io/compute=true", "--output=jsonpath={.items..metadata.name}"}
		output, err := ktests.RunCommand("oc", args...)
		Expect(err).ToNot(HaveOccurred())

		var count int
		nodes := strings.Fields(output)
		for _, node := range nodes {
			args := []string{"get", "node", node}
			output, err := ktests.RunCommand("oc", args...)
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
		})
		It("virtualmachine instance will not be re-scheduled", func() {
			By("Create the virtualmachine instance")
			args := []string{"create", "-n", ns, "-f", vmi.Manifest}
			_, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Check that the virtualmachine instance is running")
			args = []string{"get", "-n", ns, "vmi", vmi.Name, "--template", "{{.status.phase}}"}
			Eventually(func() string {
				output, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return output
			}, time.Minute*10).Should(Equal("Running"))

			By("Retrieve the vmi's node")
			args = []string{"get", "-n", ns, "vmi", vmi.Name, "--template", "{{.status.nodeName}}"}
			nodeName, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Drain the vmi's node")
			args = []string{"adm", "drain", nodeName, "--delete-local-data", "--ignore-daemonsets", "--force", "--pod-selector=kubevirt.io=virt-launcher"}
			_, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the virtual machine instance is Failed")
			args = []string{"get", "-n", ns, "vmi", vmi.Name, "--template", "{{.status.phase}}"}
			output, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal("Failed"))

			By("Verify that the vmi's node is unschedulable")
			args = []string{"get", "node", nodeName, "--template", "{{.spec.unschedulable}}"}
			output, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal("true"))

			By("Uncordon the vmi's node")
			args = []string{"adm", "uncordon", nodeName}
			_, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the node is schedulable again")
			args = []string{"get", "node", nodeName}
			output, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(ContainSubstring("SchedulingDisabled"))
		})
	})

	PDescribe("VirtualMachineInstanceReplicaSet Eviction", func() {
		BeforeEach(func() {
			vmi.Name = "vmi-replicaset-cirros"
			vmi.Manifest = "tests/manifests/vmi-replicaset-cirros.yaml"
		})
		It("virtualmachineinstance owned by a replicaset will be re-scheduled", func() {
			By("Create the virtualmachine instance replicaset")
			args := []string{"create", "-n", ns, "-f", vmi.Manifest}
			_, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Check that the virtualmachine instance is running")
			args = []string{"get", "-n", ns, "vmi", "--selector=kubevirt.io/vmReplicaSet=vmi-replicaset-cirros", "--output=jsonpath={.items..metadata.name}"}
			vmis, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			vmiList := strings.Fields(vmis)
			for _, vmi := range vmiList {
				args = []string{"get", "-n", ns, "vmi", vmi, "--template", "{{.status.phase}}"}
				Eventually(func() string {
					output, err := ktests.RunCommand("oc", args...)
					Expect(err).ToNot(HaveOccurred())
					return output
				}, time.Minute*10).Should(Equal("Running"))
			}

			By("Retrieve the vmi's node")
			args = []string{"get", "-n", ns, "vmi", vmiList[0], "--template", "{{.status.nodeName}}"}
			nodeName, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Drain the vmi's node")
			args = []string{"adm", "drain", nodeName, "--delete-local-data", "--ignore-daemonsets=true", "--force", "--pod-selector=kubevirt.io=virt-launcher"}
			_, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the vmi's node is unschedulable")
			args = []string{"get", "node", nodeName, "--template", "{{.spec.unschedulable}}"}
			Eventually(func() string {
				output, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return output
			}, time.Minute*2).Should(Equal("true"))

			By("Verify that the vmi is running on other node")
			args = []string{"get", "-n", ns, "vmi", "--selector=kubevirt.io/vmReplicaSet=vmi-replicaset-cirros", "--output=jsonpath={.items..metadata.name}"}
			vmis, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			// It's verify all VM instance are running, not just verify new VM instance status.
			vmiList = strings.Fields(vmis)
			for _, vmi := range vmiList {
				args = []string{"get", "-n", ns, "vmi", vmi, "--template", "{{.status.phase}}"}
				Eventually(func() string {
					output, err := ktests.RunCommand("oc", args...)
					Expect(err).ToNot(HaveOccurred())
					return output
				}, time.Minute*2).Should(Equal("Running"))
			}

			By("Uncordon the vmi's node")
			args = []string{"adm", "uncordon", nodeName}
			_, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the node is schedulable again")
			args = []string{"get", "node", nodeName}
			output, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(ContainSubstring("SchedulingDisabled"))
		})
	})

	Describe("VirtualMachine Eviction", func() {
		BeforeEach(func() {
			vmi.Manifest = "tests/manifests/vm-cirros.yaml"
			vmi.Name = "vm-cirros"
		})
		It("virtualmachineinstance owned by a virtual machine will be re-scheduled", func() {
			By("Create the virtualmachine instance's vm")
			args := []string{"create", "-n", ns, "-f", vmi.Manifest}
			_, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Start the virtual machine via virtctl")
			args = []string{"-n", ns, "start", vmi.Name}
			_, err = ktests.RunCommand("virtctl", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Check that the virtualmachine instance is running")
			args = []string{"get", "-n", ns, "vmi", vmi.Name, "--template", "{{.status.phase}}"}
			Eventually(func() string {
				output, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return output
			}, time.Minute*10).Should(Equal("Running"))

			By("Retrieve the vmi's node")
			args = []string{"get", "-n", ns, "vmi", vmi.Name, "--template", "{{.status.nodeName}}"}
			nodeName, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Drain the vmi's node")
			args = []string{"adm", "drain", nodeName, "--delete-local-data", "--ignore-daemonsets=true", "--force", "--pod-selector=kubevirt.io=virt-launcher"}
			_, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the vmi's node is unschedulable")
			args = []string{"get", "node", nodeName, "--template", "{{.spec.unschedulable}}"}
			Eventually(func() (string, error) {
				return ktests.RunCommand("oc", args...)
			}, time.Minute*2).Should(Equal("true"))

			By("Verify that the vmi is running on other node")
			args = []string{"get", "-n", ns, "vmi", vmi.Name, "--template", "{{.status.phase}}"}
			Eventually(func() string {
				output, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return output
			}, time.Minute*2).Should(Equal("Running"))

			By("Uncordon the vmi's node")
			args = []string{"adm", "uncordon", nodeName}
			_, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that the node is schedulable again")
			args = []string{"get", "node", nodeName}
			output, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(ContainSubstring("SchedulingDisabled"))
		})
	})
})

func uncordonNodes() {
	args := []string{"get", "node", "--selector=node-role.kubernetes.io/compute=true", "--output=jsonpath={.items..metadata.name}"}
	output, err := ktests.RunCommand("oc", args...)
	Expect(err).ToNot(HaveOccurred())

	nodes := strings.Fields(output)
	for _, node := range nodes {
		args := []string{"get", "node", node}
		output, err := ktests.RunCommand("oc", args...)
		Expect(err).ToNot(HaveOccurred())

		if strings.Contains(output, "SchedulingDisabled") {
			args = []string{"adm", "uncordon", node}
			_, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
		}
	}
}
