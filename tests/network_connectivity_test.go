package tests_test

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
)

const (
	privilegedPodYaml  = "tests/manifests/privileged-pod.yml"
	ovsVlanNet         = "tests/manifests/ovs-vlan-net.yml"
	ovsVsCtl           = "ovs-vsctl"
	bridgeName         = "br1_for_vxlan"
	ipAddCmd           = "sudo ip add a %s/24 dev %s \n"
	ipUpCmd            = "sudo ip link set up %s \n"
	privilegedTestUser = "privileged-test-user"
	noVlanPortName     = "ovs_novlan_port"
)

type NodesToIp struct {
	name string
	ip   string
}

var _ = Describe("[rfe_id:694][crit:medium][vendor:cnv-qe@redhat.com][level:component]Network Connectivity", func() {
	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	var pods *corev1.PodList
	var vmiList [2]*v1.VirtualMachineInstance
	var ovsVmsIp [2]string
	var nodesNames [2]string
	var defaultVmsIp [2]string
	var nodesToIpList [2]NodesToIp
	var ovsNodeIps [2]string
	var ipLinkCmd []string

	ktests.BeforeAll(func() {
		ktests.BeforeTestCleanup()
		ovsVmsIp = [2]string{"192.168.0.1", "192.168.0.2"}
		ovsNodeIps = [2]string{"192.168.0.3", "192.168.0.4"}
		ipLinkCmd = []string{"bash", "-c", "ip -o link show type veth | wc -l"}
		_, _, err := ktests.RunCommand("oc", "create", "serviceaccount", privilegedTestUser)
		Expect(err).ToNot(HaveOccurred())
		_, _, err = ktests.RunCommand("oc", "adm", "policy", "add-scc-to-user", "privileged", "-z", privilegedTestUser)
		Expect(err).ToNot(HaveOccurred())
		tests.CreateResourceWithFilePath(privilegedPodYaml)
		tests.CreateResourceWithFilePath(ovsVlanNet)
		nodes, err := virtClient.CoreV1().Nodes().List(metav1.ListOptions{
			LabelSelector: "node-role.kubernetes.io/compute=true",
		})
		Expect(err).ToNot(HaveOccurred())
		pods, err = virtClient.CoreV1().Pods(tests.NamespaceTestDefault).List(metav1.ListOptions{
			LabelSelector: "app=privileged-test-pod",
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(pods.Items).To(HaveLen(len(nodes.Items)))

		for i, node := range nodes.Items {
			var ip string
			nodesNames[i] = node.ObjectMeta.Name
			for _, addr := range node.Status.Addresses {
				if addr.Type == corev1.NodeInternalIP {
					ip = addr.Address
				}
			}
			nodesToIpList[i] = NodesToIp{name: nodesNames[i], ip: ip}
		}

		for i, pod := range pods.Items {
			nodeName := pod.Spec.NodeName
			podContainer := pod.Spec.Containers[0].Name
			var nextNodeIp string
			for _, nodeToIp := range nodesToIpList {
				if nodeToIp.name != nodeName {
					nextNodeIp = nodeToIp.ip
					break
				}
			}
			tests.WaitUntilResourceReadyByName(
				"pod",
				pod.Name,
				tests.NamespaceTestDefault,
				"-o=jsonpath='{.status.phase}'",
				"Running",
			)
			ktests.ExecuteCommandOnPod(
				virtClient, &pod, podContainer, []string{ovsVsCtl, "add-br", bridgeName},
			)
			ktests.ExecuteCommandOnPod(
				virtClient, &pod, podContainer, []string{
					ovsVsCtl,
					"add-port",
					bridgeName,
					"vxlan",
					"--",
					"set", "Interface", "vxlan", "type=vxlan", "options:remote_ip=" + nextNodeIp,
				},
			)
			ktests.ExecuteCommandOnPod(
				virtClient, &pod, podContainer, []string{
					ovsVsCtl,
					"add-port",
					bridgeName,
					noVlanPortName,
					"--",
					"set", "Interface", noVlanPortName, "type=internal",
				},
			)
			ktests.ExecuteCommandOnPod(
				virtClient, &pod, podContainer, []string{
					"ip", "addr", "add", ovsNodeIps[i], "dev", noVlanPortName,
				},
			)
		}

		for i := range vmiList {
			vmiList[i] = ktests.NewRandomVMIWithEphemeralDiskAndUserdata(
				ktests.ContainerDiskFor(ktests.ContainerDiskCirros), "#!/bin/bash\necho 'hello'\n")
			vmiList[i].Spec.Domain.Devices.Interfaces = []v1.Interface{
				{
					Name: "default",
					InterfaceBindingMethod: v1.InterfaceBindingMethod{
						Bridge: &v1.InterfaceBridge{},
					},
				},
				{
					Name: "vm-ovs-vlan-net",
					InterfaceBindingMethod: v1.InterfaceBindingMethod{
						Bridge: &v1.InterfaceBridge{},
					},
				},
			}
			vmiList[i].Spec.Networks = []v1.Network{
				{
					Name: "default",
					NetworkSource: v1.NetworkSource{
						Pod: &v1.PodNetwork{},
					},
				},
				{
					Name: "vm-ovs-vlan-net",
					NetworkSource: v1.NetworkSource{
						Multus: &v1.CniNetwork{
							NetworkName: "ovs-vlan-net",
						},
					},
				},
			}
			ktests.StartVmOnNode(vmiList[i], nodesNames[i])
			vmiList[i], err = virtClient.VirtualMachineInstance(
				tests.NamespaceTestDefault).Get(vmiList[i].Name, &metav1.GetOptions{})
			defaultVmsIp[i] = vmiList[i].Status.Interfaces[0].IP

			expecter, err := ktests.LoggedInCirrosExpecter(vmiList[i])
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()

			_, err = expecter.ExpectBatch([]expect.Batcher{
				&expect.BSnd{S: fmt.Sprintf(ipAddCmd, ovsVmsIp[i], "eth1")},
				&expect.BExp{R: ""},
				&expect.BSnd{S: fmt.Sprintf(ipUpCmd, "eth1")},
				&expect.BExp{R: ""},
			}, 60)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("[test_id:1609]Connectivity between VM and VM - Multus, OVS", func() {
		for i := range vmiList {
			expecter, err := ktests.LoggedInCirrosExpecter(vmiList[i])
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()
			_, err = expecter.ExpectBatch([]expect.Batcher{
				&expect.BSnd{S: "ping -w 3 " + ovsVmsIp[1-i] + "\n"},
				&expect.BExp{R: "3 packets transmitted"},
				&expect.BSnd{S: "echo $?\n"},
				&expect.BExp{R: "0"},
			}, 30*time.Second)
			Expect(err).ToNot(HaveOccurred())
		}
	})
	It("[test_id:1610]Connectivity between VM and VM", func() {
		for i := range vmiList {
			expecter, err := ktests.LoggedInCirrosExpecter(vmiList[i])
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()
			_, err = expecter.ExpectBatch([]expect.Batcher{
				&expect.BSnd{S: "ping -w 3 " + defaultVmsIp[1-i] + "\n"},
				&expect.BExp{R: "3 packets transmitted"},
				&expect.BSnd{S: "echo $?\n"},
				&expect.BExp{R: "0"},
			}, 30*time.Second)
			Expect(err).ToNot(HaveOccurred())
		}
	})
	It("[test_id:743]Connection should failed between no VLAN specified interface and VM with VLAN network", func() {
		expecter, err := ktests.LoggedInCirrosExpecter(vmiList[0])
		Expect(err).ToNot(HaveOccurred())
		defer expecter.Close()
		_, err = expecter.ExpectBatch([]expect.Batcher{
			&expect.BSnd{S: "ping -w 3 " + ovsNodeIps[0] + "\n"},
			&expect.BExp{R: "3 packets transmitted"},
			&expect.BSnd{S: "echo $?\n"},
			&expect.BExp{R: "1"},
		}, 30*time.Second)
		Expect(err).ToNot(HaveOccurred())
	})
	It("[test_id:681]The veth will be removed after deleting the VM", func() {
		// Make sure this test excute last since the VMs are removed in the test
		var out string
		var numberOfVethsBeforeDelete [2]int
		var numberOfVethsAfterDelete [2]int
		// Get number of veth for each node while VMs are running
		for i, pod := range pods.Items {
			podContainer := pod.Spec.Containers[0].Name
			out, err = ktests.ExecuteCommandOnPod(
				virtClient, &pod, podContainer, ipLinkCmd,
			)
			Expect(err).ToNot(HaveOccurred())
			stripOut := strings.TrimSuffix(out, "\n")
			intOut, err := strconv.Atoi(stripOut)
			Expect(err).ToNot(HaveOccurred())
			numberOfVethsBeforeDelete[i] = intOut
		}
		By("Delete VMs")
		for _, vm := range vmiList {
			tests.DeleteResourceByName("vmi", tests.NamespaceTestDefault, vm.ObjectMeta.Name)
		}
		// Get number of veth for each node after VMs was deleted
		for i, pod := range pods.Items {
			podContainer := pod.Spec.Containers[0].Name
			out, err := ktests.ExecuteCommandOnPod(
				virtClient, &pod, podContainer, ipLinkCmd,
			)
			Expect(err).ToNot(HaveOccurred())
			stripOut := strings.TrimSuffix(out, "\n")
			intOut, err := strconv.Atoi(stripOut)
			Expect(err).ToNot(HaveOccurred())
			numberOfVethsAfterDelete[i] = intOut
		}
		// Check that we have 2 veth less for each node (each VM have 2 interfaces)
		By("Chack that all veth interfaces that was used by the VMs are deleted from the nodes")
		for i := range numberOfVethsBeforeDelete {
			Expect(numberOfVethsBeforeDelete[i] - 2).To(BeEquivalentTo(numberOfVethsAfterDelete[i]))
		}
	})
})
