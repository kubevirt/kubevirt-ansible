package tests_test

import (
	"flag"
	"fmt"
	"strings"
	"time"

	expect "github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/log"
	hw_utils "kubevirt.io/kubevirt/pkg/util/hardware"
	ktests "kubevirt.io/kubevirt/tests"
)

var _ = Describe("Configurations", func() {

	flag.Parse()

	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	Describe("VirtualMachineInstance with CPU pinning", func() {
		var nodes *kubev1.NodeList
		BeforeEach(func() {
			nodes, err = virtClient.CoreV1().Nodes().List(metav1.ListOptions{})
			ktests.PanicOnError(err)
		})

		Context("with cpu pinning enabled", func() {
			It("non master node should have a cpumanager label", func() {
				cpuManagerEnabled := false
				for idx := 1; idx < len(nodes.Items); idx++ {
					labels := nodes.Items[idx].GetLabels()
					for label, val := range labels {
						if label == "cpumanager" && val == "true" {
							cpuManagerEnabled = true
						}
					}
				}
				Expect(cpuManagerEnabled).To(BeTrue())
			})

			It("Should start a vm with cpu pinning using the spec.domain.resources.cpu ", func() {

				cpuVmi := ktests.NewRandomVMIWithEphemeralDiskAndUserdata(
					ktests.ContainerDiskFor(
						ktests.ContainerDiskFedora), "#!/bin/bash\necho \"fedora\" | passwd fedora --stdin\n")
				cpuVmi.Spec.Domain.CPU = &v1.CPU{
					DedicatedCPUPlacement: true,
				}
				cpuVmi.Spec.Domain.Resources = v1.ResourceRequirements{
					Requests: kubev1.ResourceList{
						kubev1.ResourceCPU:    resource.MustParse("2"),
						kubev1.ResourceMemory: resource.MustParse("512M"),
					},
				}

				By("Starting a VirtualMachineInstance")
				_, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(cpuVmi)
				Expect(err).ToNot(HaveOccurred())
				node := ktests.WaitForSuccessfulVMIStart(cpuVmi)
				Expect(node).To(ContainSubstring("node"))

				By("Checking that the pod QOS is guaranteed")
				readyPod := ktests.GetRunningPodByVirtualMachineInstance(cpuVmi, tests.NamespaceTestDefault)
				podQos := readyPod.Status.QOSClass
				Expect(podQos).To(Equal(kubev1.PodQOSGuaranteed))

				var computeContainer *kubev1.Container
				for _, container := range readyPod.Spec.Containers {
					if container.Name == "compute" {
						computeContainer = &container
					}
				}
				if computeContainer == nil {
					ktests.PanicOnError(fmt.Errorf("could not find the compute container"))
				}

				output, err := ktests.ExecuteCommandOnPod(
					virtClient,
					readyPod,
					"compute",
					[]string{"cat", hw_utils.CPUSET_PATH},
				)
				log.Log.Infof("%v", output)
				By("Expecting the cpu count right from the cgroups")
				Expect(err).ToNot(HaveOccurred())
				output = strings.TrimSuffix(output, "\n")
				pinnedCPUsList, err := hw_utils.ParseCPUSetLine(output)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(pinnedCPUsList)).To(Equal(2))

				By("Expecting the VirtualMachineInstance console")
				expecter, err := tests.LoggedInFedoraExpecter(cpuVmi.Name, tests.NamespaceTestDefault, 360)
				Expect(err).ToNot(HaveOccurred())
				defer expecter.Close()

				By("Checking the number of CPU cores under guest OS")
				res, err := expecter.ExpectBatch([]expect.Batcher{
					&expect.BSnd{S: "grep -c ^processor /proc/cpuinfo\n"},
					&expect.BExp{R: "2"},
				}, 15*time.Second)
				log.DefaultLogger().Object(cpuVmi).Infof("%v", res)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("VM with both dedicated cpu and non dedicated cpu should be possible on same node", func() {

			var cpuvmi, vmi *v1.VirtualMachineInstance

			BeforeEach(func() {

				nodes := ktests.GetAllSchedulableNodes(virtClient)
				Expect(nodes.Items).ToNot(BeEmpty(), "There should be some nodes")
				node := nodes.Items[1].Name

				vmi = ktests.NewRandomVMIWithEphemeralDiskAndUserdata(
					ktests.ContainerDiskFor(
						ktests.ContainerDiskFedora), "#!/bin/bash\necho \"fedora\" | passwd fedora --stdin\n")
				cpuvmi = ktests.NewRandomVMIWithEphemeralDiskAndUserdata(
					ktests.ContainerDiskFor(
						ktests.ContainerDiskFedora), "#!/bin/bash\necho \"fedora\" | passwd fedora --stdin\n")
				cpuvmi.Spec.Domain.CPU = &v1.CPU{
					Cores: 1,
					DedicatedCPUPlacement: true,
				}
				cpuvmi.Spec.Domain.Resources = v1.ResourceRequirements{
					Requests: kubev1.ResourceList{
						kubev1.ResourceMemory: resource.MustParse("512M"),
					},
				}
				cpuvmi.Spec.NodeSelector = map[string]string{"kubernetes.io/hostname": node}

				vmi.Spec.Domain.CPU = &v1.CPU{
					Cores: 1,
				}
				vmi.Spec.Domain.Resources = v1.ResourceRequirements{
					Requests: kubev1.ResourceList{
						kubev1.ResourceMemory: resource.MustParse("512M"),
					},
				}
				vmi.Spec.NodeSelector = map[string]string{"kubernetes.io/hostname": node}
			})

			It("should start a vm with no cpu pinning after a vm with cpu pinning on same node", func() {

				By("Starting a VirtualMachineInstance with dedicated cpus")
				_, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(cpuvmi)
				Expect(err).ToNot(HaveOccurred())
				node1 := ktests.WaitForSuccessfulVMIStart(cpuvmi)
				Expect(node1).To(ContainSubstring("node2"))

				By("Starting a VirtualMachineInstance without dedicated cpus")
				_, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
				Expect(err).ToNot(HaveOccurred())
				node2 := ktests.WaitForSuccessfulVMIStart(vmi)
				Expect(node2).To(ContainSubstring("node2"))
			})

			It("should start a vm with cpu pinning after a vm with no cpu pinning on same node", func() {

				By("Starting a VirtualMachineInstance without dedicated cpus")
				_, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
				Expect(err).ToNot(HaveOccurred())
				node2 := ktests.WaitForSuccessfulVMIStart(vmi)
				Expect(node2).To(ContainSubstring("node2"))

				By("Starting a VirtualMachineInstance with dedicated cpus")
				_, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(cpuvmi)
				Expect(err).ToNot(HaveOccurred())
				node1 := ktests.WaitForSuccessfulVMIStart(cpuvmi)
				Expect(node1).To(ContainSubstring("node2"))
			})
		})
	})
})
