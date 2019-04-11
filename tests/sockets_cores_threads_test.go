/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2019 Red Hat, Inc.
 *
 */

package tests_test

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	tframework "kubevirt.io/kubevirt-ansible/tests/framework"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
	ktests "kubevirt.io/kubevirt/tests"
)

func parseYAMLConfig(podYAML string) (bool, int, int) {
	isCPUPresent := false
	var resourcesRequests int64 = 0
	var resourcesLimits int64 = 0

	var config corev1.Pod
	buf := bytes.NewBuffer([]byte(podYAML))
	decoder := k8syaml.NewYAMLOrJSONDecoder(buf, 1024)
	decoder.Decode(&config)

	// Checking that fields Requests and Limits exist
	if len(config.Spec.Containers[0].Resources.Requests) != 0 && len(config.Spec.Containers[0].Resources.Limits) != 0 {

		element, doesElementExist := config.Spec.Containers[0].Resources.Requests["cpu"]
		if doesElementExist {
			isCPUPresent = true
			resourcesRequests, _ = (&element).AsInt64()
		}

		element, doesElementExist = config.Spec.Containers[0].Resources.Limits["cpu"]
		if doesElementExist {
			isCPUPresent = true
			resourcesLimits, _ = (&element).AsInt64()
		}
	}

	return isCPUPresent, int(resourcesRequests), int(resourcesLimits)
}

// Checking if the cluster can run at least one VM
func isEnoughResources(virtClient kubecli.KubevirtClient, cpuNeeded int, memNeeded int64) (bool, int, int) {
	availableVMs, cpu_limit, mem_limit := tframework.GetAvailableResources(virtClient, int64(cpuNeeded), memNeeded)
	if availableVMs == 0 {
		return false, cpu_limit, mem_limit

	} else {
		return true, cpu_limit, mem_limit

	}
}

func getYAMLFilename(sockets, cores, threads int, address_common string) string {
	// 0 means parameter set to 0, 1 means parameter set to non-zero
	var file_name [2][2][2]string
	file_name[0][0][0] = "vm-template-cirros-no-sockets-cores-and-threads.yaml"
	file_name[1][0][0] = "vm-template-cirros-only-sockets.yaml"
	file_name[0][1][0] = "vm-template-cirros-only-cores.yaml"
	file_name[0][0][1] = "vm-template-cirros-only-threads.yaml"
	file_name[0][1][1] = "vm-template-cirros-only-cores-and-threads.yaml"
	file_name[1][0][1] = "vm-template-cirros-only-sockets-and-threads.yaml"
	file_name[1][1][0] = "vm-template-cirros-only-sockets-and-cores.yaml"
	file_name[1][1][1] = "vm-template-cirros-sockets-cores-and-threads.yaml"

	// Because go doesn't have ternary operators
	s := 0
	if sockets > 0 {
		s = 1
	}

	c := 0
	if cores > 0 {
		c = 1
	}

	t := 0
	if threads > 0 {
		t = 1
	}

	return address_common + file_name[s][c][t]
}

func getArguments(vm_name string, sockets int, cores int, threads int) []string {
	arguments := []string{"NAME=" + vm_name}

	if sockets != 0 {
		arguments = append(arguments, "CPU_SOCKETS="+strconv.Itoa(sockets))
	}
	if cores != 0 {
		arguments = append(arguments, "CPU_CORES="+strconv.Itoa(cores))
	}
	if threads != 0 {
		arguments = append(arguments, "CPU_THREADS="+strconv.Itoa(threads))
	}

	return arguments
}

func clean_pods(virtClient kubecli.KubevirtClient, requiredPods []*corev1.Pod) {
	listOptions := metav1.ListOptions{}
	podList, err := virtClient.CoreV1().Pods(ktests.NamespaceTestDefault).List(listOptions)

	Expect(err).ToNot(HaveOccurred())
	for _, item := range podList.Items {

		deletePod := true
		for _, internalItem := range requiredPods {
			fmt.Println("internalItem=", internalItem.Name)
			if item.Name == internalItem.Name {
				deletePod = false
				fmt.Println("do not delete pod=", internalItem.Name)
			}

		}
		if deletePod {
			fmt.Println("delete pod=", item.Name)
			_, _, _ = ktests.RunCommandWithNS(ktests.NamespaceTestDefault, "oc", "delete", "pod", item.Name)
		}
	}
}

var _ = Describe("[rfe_id:1443][crit:medium]vendor:cnv-qe@redhat.com][level:component]Check CPU topology inside VM", func() {
	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)
	address_common := "tests/manifests/sockets_cores_threads/"
	var vmi11, vmi121, vmi122, vmi123 tframework.VirtualMachine
	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	Context("test case 1.1 Check the validity of the XML file if user didn’t set the CPU topology at all", func() {
		It("[test_id:1485] testcase 1.1 Check the validity of the XML file if user didn’t set the CPU topology at all", func() {
			//It("Pre-checks"",func() {
			vmi11.Manifest = address_common + "vmi-case1.1.yml"
			vmi11.Namespace = ktests.NamespaceTestDefault

			By("Creating VMI using manifest")
			_, _, err := vmi11.Create()
			Expect(err).ToNot(HaveOccurred())
			vmi11.Name = "vmi-case1.1"

			By("Getting VMI object")
			getVMOptions := metav1.GetOptions{}
			vmi, err := virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Get(vmi11.Name, &getVMOptions)
			Expect(err).ToNot(HaveOccurred())

			By("Waiting until VMI start")
			ktests.WaitForSuccessfulVMIStart(vmi)

			By("getting pod object")
			vmiPod := ktests.GetRunningPodByVirtualMachineInstance(vmi, ktests.NamespaceTestDefault)

			By("clean old pods")
			var requiredPods []*corev1.Pod
			requiredPods = append(requiredPods, vmiPod)
			clean_pods(virtClient, requiredPods)

			By("Checking resources in the pod")
			podName := vmiPod.Name
			outPodYAML, _, err := ktests.RunCommandWithNS(ktests.NamespaceTestDefault, "oc", "get", "pod", podName, "-o", "yaml")
			Expect(err).ToNot(HaveOccurred())
			cpuExist, _, _ := parseYAMLConfig(outPodYAML)

			Expect(cpuExist).To(BeFalse())

			By("Checking XML")
			vmiXml, err := ktests.GetRunningVirtualMachineInstanceDomainXML(
				virtClient,
				vmi,
			)

			domStat := &api.DomainSpec{}
			err = xml.Unmarshal([]byte(vmiXml), domStat)
			Expect(err).ToNot(HaveOccurred())
			Expect(uint(domStat.VCPU.CPUs) == 1).To(BeTrue())

			Expect(uint(domStat.CPU.Topology.Sockets) == 1).To(BeTrue())
			Expect(uint(domStat.CPU.Topology.Cores) == 1).To(BeTrue())
			Expect(uint(domStat.CPU.Topology.Threads) == 1).To(BeTrue())

		})
	})

	Context("test case 1.2 Check the validity of the XML file if user didn’t set the CPU topology (only standard cpus - previous releases like) ", func() {
		It("[test_id:1487]testcase 1.2 Check the validity of the XML file if user didn’t set the CPU topology (only standard cpus - previous releases like)  ", func() {

			By("declare variables")
			vmi121.Manifest = address_common + "/vmi-case1.2.1.yml"
			vmi122.Manifest = address_common + "/vmi-case1.2.2.yml"
			vmi123.Manifest = address_common + "/vmi-case1.2.3.yml"
			vmi121.Name = "vmi-case1.2.1"
			vmi122.Name = "vmi-case1.2.2"
			vmi123.Name = "vmi-case1.2.3"
			vmi121.Namespace, vmi122.Namespace, vmi123.Namespace = ktests.NamespaceTestDefault, ktests.NamespaceTestDefault, ktests.NamespaceTestDefault

			By("Create VMI using manifest - only request CPU resources - VMI-2.1.1")
			_, _, err := vmi121.Create()
			Expect(err).ToNot(HaveOccurred())

			By("Create VMI using manifest - request and limit cpu resources, same amount - VMI-2.2.2")
			_, _, err = vmi122.Create()
			Expect(err).ToNot(HaveOccurred())

			By("create VMI using wrong manifest, request and limit has different cpu resources")
			_, _, err = vmi123.Create()
			Expect(err).Should(HaveOccurred())

			By("Getting VMI object")
			getVMOptions := metav1.GetOptions{}
			vmi121, err := virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Get(vmi121.Name, &getVMOptions)
			Expect(err).ToNot(HaveOccurred())
			vmi122, err := virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Get(vmi122.Name, &getVMOptions)
			Expect(err).ToNot(HaveOccurred())

			By("Waiting until VMI121 start")
			ktests.WaitForSuccessfulVMIStart(vmi121)

			By("Waiting until VMI122 start")
			ktests.WaitForSuccessfulVMIStart(vmi122)

			By("getting pods objects")
			vmi121Pod := ktests.GetRunningPodByVirtualMachineInstance(vmi121, ktests.NamespaceTestDefault)
			vmi122Pod := ktests.GetRunningPodByVirtualMachineInstance(vmi122, ktests.NamespaceTestDefault)

			By("workaround for CI - clean old pods")
			var requiredPods []*corev1.Pod
			requiredPods = append(requiredPods, vmi121Pod)
			requiredPods = append(requiredPods, vmi122Pod)
			fmt.Println("requiredPods[0].Name=", requiredPods[0].Name, " requiredPods[1].Name=", requiredPods[1].Name)
			clean_pods(virtClient, requiredPods)

			By("Checking that pod was created and has the right name")
			Expect(err).ToNot(HaveOccurred())
			Expect(vmi121Pod.Name).To(HavePrefix("virt-launcher-"+vmi121.Name), "Pod's name should contain name of the VM associated with it")
			Expect(vmi122Pod.Name).To(HavePrefix("virt-launcher-"+vmi122.Name), "Pod's name should contain name of the VM associated with it")

			By("get pod-s YAML files")
			podName121 := vmi121Pod.Name
			podName122 := vmi122Pod.Name
			outPodYAML121, _, err := ktests.RunCommandWithNS(ktests.NamespaceTestDefault, "oc", "get", "pod", podName121, "-o", "yaml")
			Expect(err).ToNot(HaveOccurred())
			outPodYAML122, _, err := ktests.RunCommandWithNS(ktests.NamespaceTestDefault, "oc", "get", "pod", podName122, "-o", "yaml")
			Expect(err).ToNot(HaveOccurred())

			By("Checking resources in the pod")
			var CPUresourcesRequests, CPUresourcesLimits int
			cpuExist, CPUresourcesRequests, CPUresourcesLimits := parseYAMLConfig(outPodYAML121)
			Expect(cpuExist).To(BeTrue())
			Expect(CPUresourcesRequests == 2).To(BeTrue())
			Expect(CPUresourcesLimits == 0).To(BeTrue())
			cpuExist, CPUresourcesRequests, CPUresourcesLimits = parseYAMLConfig(outPodYAML122)
			Expect(cpuExist).To(BeTrue())
			Expect(CPUresourcesRequests == 2).To(BeTrue())
			Expect(CPUresourcesLimits == 2).To(BeTrue())

			By("Get vmi121 XML")
			vmiXml121, err := ktests.GetRunningVirtualMachineInstanceDomainXML(
				virtClient,
				vmi121,
			)

			By("Get vmi122 XML")
			vmiXml122, err := ktests.GetRunningVirtualMachineInstanceDomainXML(
				virtClient,
				vmi122,
			)

			By("vmi unmarshal")
			domStat121 := &api.DomainSpec{}
			domStat122 := &api.DomainSpec{}
			err = xml.Unmarshal([]byte(vmiXml121), domStat121)
			Expect(err).ToNot(HaveOccurred())

			err = xml.Unmarshal([]byte(vmiXml122), domStat122)
			Expect(err).ToNot(HaveOccurred())

			By("check XML")

			By("check XML - vmi121")
			Expect(uint(domStat121.VCPU.CPUs) == 2).To(BeTrue())
			Expect(uint(domStat121.CPU.Topology.Sockets) == 2).To(BeTrue())
			Expect(uint(domStat121.CPU.Topology.Cores) == 1).To(BeTrue())
			Expect(uint(domStat121.CPU.Topology.Threads) == 1).To(BeTrue())

			By("check XML - vmi122")
			Expect(uint(domStat122.VCPU.CPUs) == 2).To(BeTrue())
			Expect(uint(domStat122.CPU.Topology.Sockets) == 2).To(BeTrue())
			Expect(uint(domStat122.CPU.Topology.Cores) == 1).To(BeTrue())
			Expect(uint(domStat122.CPU.Topology.Threads) == 1).To(BeTrue())

			By("Clean old pods")
			var cleanPods []*corev1.Pod
			clean_pods(virtClient, cleanPods)

		})
	})

	Context("test cases 1.3,2.1,2.2 Check the validity of the XML file if user didn’t set the CPU topology at all", func() {
		It("[test_id:1488] [test_id:1490] [test_id:1489] [test_id:1488] testcases 1.3,2.1,2.2 Check the validity of the XML file if user didn’t set the CPU topology at all", func() {

			var wg sync.WaitGroup
			vm_index := 0

			for sockets := 0; sockets < 3; sockets++ {
				for cores := 0; cores < 3; cores++ {
					for threads := 0; threads < 3; threads++ {

						By("1.3 check resources in the cluster")
						cpuNeeded := 1
						if sockets > 0 {
							cpuNeeded *= sockets
						}
						if cores > 0 {
							cpuNeeded *= cores
						}
						if threads > 0 {
							cpuNeeded *= threads
						}
						const memNeeded int64 = 256 * 1024 * 1024 // 256mb is default in the template

						isAvailable := false
						const maxWaitIterations = 15 // half of minute
						var cpuLimit, memLimit int
						var IsResourcesInCluster bool
						for i := 0; i < maxWaitIterations; i++ {
							IsResourcesInCluster, cpuLimit, memLimit = isEnoughResources(virtClient, cpuNeeded, memNeeded)
							if IsResourcesInCluster {
								isAvailable = true
								break
							}
							time.Sleep(2 * time.Second)
						}
						//workaroud for current resources in CI
						if cpuLimit >= 1 && memLimit >= 1 {
							Expect(isAvailable).To(BeTrue(), "Cluster should have enough resources")
						} else {
							break
						}

						vm_name := "vm13-" + strconv.Itoa(vm_index)
						vm_index++

						filename := getYAMLFilename(sockets, cores, threads, address_common)
						arguments := getArguments(vm_name, sockets, cores, threads)

						virtRawVMFilePath := address_common + "/sockets_cores_threads_raw_manifest_" + vm_name + ".yaml"
						tframework.ProcessTemplateWithParameters(filename, virtRawVMFilePath, arguments...)

						By("1.3 Starting gouroutine to create, start and test VM")
						wg.Add(1)
						//prevent server overloading - CI workaround (i/o error)
						time.Sleep(10 * time.Second)

						go func(sockets int, cores int, threads int, wg *sync.WaitGroup) {
							defer wg.Done()
							By("1.3 Create VM from template and launch it")

							tframework.CreateResourceWithFilePathTestNamespace(virtRawVMFilePath)
							_, _, err := ktests.RunCommandWithNS(ktests.NamespaceTestDefault, "virtctl", "start", vm_name)
							Expect(err).ToNot(HaveOccurred())
							_, _, err = ktests.RunCommandWithNS(ktests.NamespaceTestDefault, "oc", "project", ktests.NamespaceTestDefault)
							Expect(err).ToNot(HaveOccurred())

							By("1.3 Getting VMI object")
							getVMOptions := metav1.GetOptions{}
							vmi, err := virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Get(vm_name, &getVMOptions)
							Expect(err).ToNot(HaveOccurred())
							ktests.WaitForSuccessfulVMIStart(vmi)

							By("1.3 Checking that pod was created and has the right name")
							vmiPod_vmNumName := ktests.GetRunningPodByVirtualMachineInstance(vmi, ktests.NamespaceTestDefault)
							podName := vmiPod_vmNumName.Name
							Expect(podName).To(HavePrefix("virt-launcher-"+vm_name), "Pod's name should contain name of the VM associated with it")

							By("1.3 Checking resources in the pod")
							outPodYAML, _, err := ktests.RunCommandWithNS(ktests.NamespaceTestDefault, "oc", "get", "pod", podName, "-o", "yaml")
							Expect(err).ToNot(HaveOccurred())
							cpuExist, _, _ := parseYAMLConfig(outPodYAML)
							Expect(cpuExist).To(BeFalse(), "YAML should have CPUs")

							By("1.3 Get VMI XML")
							vmiXml, err := ktests.GetRunningVirtualMachineInstanceDomainXML(virtClient, vmi)
							Expect(err).ToNot(HaveOccurred())

							By("1.3 VMI Unmarshal")
							domStat := &api.DomainSpec{}
							err = xml.Unmarshal([]byte(vmiXml), domStat)
							Expect(err).ToNot(HaveOccurred())

							XMLSockets := sockets
							XMLCores := cores
							XMLThreads := threads
							// If CPU cores, sockets or threads set to 0, XML should have 1 for this parameter
							if sockets == 0 {
								XMLSockets = 1
							}
							if cores == 0 {
								XMLCores = 1
							}
							if threads == 0 {
								XMLThreads = 1
							}

							By("1.3 Checking XML topology")
							Expect(int(domStat.CPU.Topology.Sockets) == XMLSockets).To(BeTrue(), "XML should have right number of sockets")
							Expect(int(domStat.CPU.Topology.Cores) == XMLCores).To(BeTrue(), "XML should have right number of cores")
							Expect(int(domStat.CPU.Topology.Threads) == XMLThreads).To(BeTrue(), "XML should have right number of threads")

							By("1.3 Checking the amount of vCPU")
							vCPUAmount := XMLSockets * XMLCores * XMLThreads
							Expect(int(domStat.VCPU.CPUs) == vCPUAmount).To(BeTrue(), "XML should have right number of vCPUs")

							// TC 2.1 and 2.2 should do the same as 1.3 but with several additional checks at the end.
							// Creating and destroying all these VMs second time may be unnecessary time consuming
							// TODO: 2.1 & 2.2 - move it to independent test case?

							By("2.1 Expecting the VirtualMachineInstance console")
							expecter, err := ktests.LoggedInCirrosExpecter(vmi)
							Expect(err).ToNot(HaveOccurred(), "Console should be started")
							defer expecter.Close()

							By("2.2 Checking the number of sockets in guest OS")
							_, err = expecter.ExpectBatch([]expect.Batcher{
								&expect.BSnd{S: "lscpu | grep Socket | awk '{print $2}'\n"},
								&expect.BExp{R: strconv.Itoa(XMLSockets)},
							}, 60*time.Second)
							Expect(err).ToNot(HaveOccurred(), "Should report number of sockets")

							/*By("2.2 Checking the number of cores in guest OS")
							_, err = expecter.ExpectBatch([]expect.Batcher{
								&expect.BSnd{S: "lscpu | grep Core | awk '{print $4}'\n"},
								&expect.BExp{R: strconv.Itoa(XMLCores)},
							}, 60*time.Second)
							Expect(err).ToNot(HaveOccurred(), "Should report number of cores")

							By("2.2 Checking the number of threads in guest OS")
							_, err = expecter.ExpectBatch([]expect.Batcher{
								&expect.BSnd{S: "lscpu | grep Thread | awk '{print $4}'\n"},
								&expect.BExp{R: strconv.Itoa(XMLThreads)},
							}, 60*time.Second)
							Expect(err).ToNot(HaveOccurred(), "Should report number of threads")*/

							By("Deleting VM")
							_, _, _ = ktests.RunCommandWithNS(ktests.NamespaceTestDefault, "oc", "delete", "vm", vm_name)
							Expect(err).ToNot(HaveOccurred())
							By("Deleting VM manifest")
							err = os.Remove(virtRawVMFilePath)
							Expect(err).ToNot(HaveOccurred())

						}(sockets, cores, threads, &wg)
					}
				}
			}
			wg.Wait()
		})
	})

})
