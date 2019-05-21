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
	"flag"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"

	"encoding/json"
	"encoding/xml"
	"github.com/mholt/archiver"
	"io/ioutil"
	k8sv1 "k8s.io/api/core/v1"
	tframework "kubevirt.io/kubevirt-ansible/tests/framework"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
	ktests "kubevirt.io/kubevirt/tests"
	"reflect"
	"time"
)

const (
	windowsDisk   = "windows-disk"
	username      = "Administrator"
	password      = "Heslo123"
	winrmCli      = "winrmcli"
	winrmCliCmd   = "winrm-cli"
	winrmCp       = "/usr/bin/winrmcp"
	winrmCliImage = "polnoch/winrmcliwinrmcp"
)

type hvinfoJson struct {
	HyperVsupport bool `json:"HyperVsupport"`
}

func getArgument(vm_name string, add_pvc bool) []string {
	argument := []string{"NAME=" + vm_name}
	if add_pvc {
		argument = append(argument, "PVCNAME="+vm_name+"-pvc")
	}
	return argument
}

func isVmiStarted(vmi tframework.VirtualMachine) bool {
	vmi.Type = "vmi"
	const maxWaitIterations = 60 // two minutes
	for i := 0; i < maxWaitIterations; i++ {
		vmRunning, err := vmi.IsRunning()
		Expect(err).ToNot(HaveOccurred())
		if vmRunning {
			fmt.Println("VM has launched")
			return true
		}

		fmt.Println("Check launching process, attempt No=", i)
		time.Sleep(2 * time.Second)
	}

	return false
}

func waitUntilWindowsVMHasBoot(ip, username, password string, virtClient kubecli.KubevirtClient, winrmcliPod *k8sv1.Pod) bool {
	const maxWaitIterations = 60 // two minutes
	const testCommand = "wmic csproduct get \"UUID\""
	for i := 0; i < maxWaitIterations; i++ {
		commandOutput, resultCommand := tframework.RunPsCommandInWindowsVM(winrmCliCmd, ip, username, password, testCommand, virtClient, winrmcliPod)
		if resultCommand {
			fmt.Println("BREAK! resultCommand=", resultCommand, " i=", i)
			return true
		}
		fmt.Println("resultCommand", resultCommand, "commandOutput=", commandOutput, "i=", i)
		time.Sleep(2 * time.Second)
	}

	return false
}

var _ = Describe("windowsTest", func() {

	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)
	address_common := "tests/manifests/"
	BeforeEach(func() {

	})

	Context("check hyperv", func() {
		It("[test_id:2170] HyperV feature - checking XML", func() {

			var vmiHyperv tframework.VirtualMachine
			vmiHyperv.Manifest = address_common + "/hyperv.yml"
			vmiHyperv.Namespace = ktests.NamespaceTestDefault

			By("Creating VMI using manifest")
			_, _, err := vmiHyperv.Create()
			Expect(err).ToNot(HaveOccurred())

			vmiHyperv.Name = "hyperv-test"

			By("Getting VMI object")
			getVMOptions := metav1.GetOptions{}
			vmi, err := virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Get(vmiHyperv.Name, &getVMOptions)
			Expect(err).ToNot(HaveOccurred())

			By("Waiting until VMI start")
			ktests.WaitForSuccessfulVMIStart(vmi)

			By("Get vmi XML")
			vmiXml, err := ktests.GetRunningVirtualMachineInstanceDomainXML(
				virtClient,
				vmi,
			)

			By("unmarshal XML")
			domStat := &api.DomainSpec{}
			err = xml.Unmarshal([]byte(vmiXml), domStat)
			Expect(err).ToNot(HaveOccurred())

			By("Checking HyperV container features for KVM")
			features := domStat.Features
			Expect(features).ToNot(BeNil(), "features container can't be nil")

			By("Checking HyperV container in the XML")
			hyperv := features.Hyperv
			Expect(hyperv).ToNot(BeNil(), "hyperv container can't be nil")

			Expect((*hyperv.Relaxed).State).To(Equal("on"))
			Expect((*hyperv.VAPIC).State).To(Equal("on"))
			Expect((*hyperv.Spinlocks).State).To(Equal("on"))
			Expect(*hyperv.Spinlocks.Retries).To(Equal(uint32(8191)))

			By("Checking HyperV timer emulation")
			Timer := domStat.Clock.Timer

			for _, t := range Timer {
				fmt.Println(t.Name)
				switch t.Name {
				case "hpet":
					Expect(t.Present).To(Equal("no"))

				case "hypervclock":
					Expect(t.Present).To(Equal("yes"))
				}
			}

		})
	})

	Context("check hyperv", func() {
		It("[test_id:2171] HyperV feature - checking inside VM", func() {

			By("Creating winrm-cli pod for the future use")
			winrmCliPod := tframework.CreatingWinRmiPod(winrmCli, virtClient, true, winrmCliImage)

			By("Declaring VM")
			var vm tframework.VirtualMachine
			vm.Manifest = address_common + "/hyperv2.yml"
			vm.Name = "hyperv-test2"
			vm.Namespace = ktests.NamespaceTestDefault

			By("Creating VM using manifest")
			_, _, err := vm.Create()
			Expect(err).ToNot(HaveOccurred())

			By("Declare PVC")
			virtRawPvcPath := "_out/windows_pvc_raw_manifest_hyperv2.yaml"
			filenamePath := address_common + "mypvc_nfs.yml"
			arguments := getArgument(vm.Name, false)
			tframework.ProcessTemplateWithParameters(filenamePath, virtRawPvcPath, arguments...)
			_, _, err = ktests.RunCommandWithNS(tframework.NamespaceTestDefault, "oc", "apply", "-f", virtRawPvcPath)
			Expect(err).ToNot(HaveOccurred())

			By("Launch VM")
			_, _, err = vm.Start()
			Expect(err).ToNot(HaveOccurred())

			By("Waiting until VM starts")
			vmiWinResultStart := isVmiStarted(vm)
			Expect(vmiWinResultStart).To(BeTrue())

			By("Getting VMI object")
			getVMOptions := metav1.GetOptions{}
			vmi, err := virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Get(vm.Name, &getVMOptions)
			Expect(err).ToNot(HaveOccurred())

			By("Waiting until VMI start")
			ktests.WaitForSuccessfulVMIStart(vmi)
			Expect(err).ToNot(HaveOccurred())

			By("Getting VMI IP adress")
			ip := vmi.Status.Interfaces[0].IP

			By("Getting VMI booting")
			windowsBootResult := waitUntilWindowsVMHasBoot(ip, username, password, virtClient, winrmCliPod)
			Expect(windowsBootResult).To(BeTrue())

			By("download zip")
			hvinfoZipPath := "_out" + "/hvinfo.zip"
			url := tframework.GetLatestGitHubReleaseURL("sHaggYcaT", "hvinfo")
			hvinfoZip := tframework.DownloadFile(url)
			err = ioutil.WriteFile(hvinfoZipPath, hvinfoZip, 0644)
			Expect(err).ToNot(HaveOccurred())

			By("extract zip")
			binaryPathInsideZip := "hvinfo-0.1.0/binaries/hvinfo.exe"
			unpackedBinaryPath := "_out" + "/hvinfo"
			err = archiver.Unarchive(hvinfoZipPath, unpackedBinaryPath)
			Expect(err).ToNot(HaveOccurred())
			pathToExe := unpackedBinaryPath + "/" + binaryPathInsideZip

			By("copy binary into winrmpod")
			winRmPodPath := winrmCliPod.Name + ":" + "/tmp"
			winRmInPodPath := "/tmp/hvinfo.exe"
			winRmInWinVmPath := "C:\\hvinfo.exe"
			_, _, err = ktests.RunCommandWithNS(tframework.NamespaceTestDefault, "oc", "cp", pathToExe, winRmPodPath)
			Expect(err).ToNot(HaveOccurred())

			By("Copy binary into VM")
			tframework.CopyFileIntoWindowsVM(winrmCp, ip, username, password, winRmInPodPath, winRmInWinVmPath, virtClient, winrmCliPod)
			commandOutput, _ := tframework.RunPsCommandInWindowsVM(winrmCliCmd, ip, username, password, winRmInWinVmPath, virtClient, winrmCliPod)
			hvinfoStruct := &hvinfoJson{}
			err = json.Unmarshal([]byte(commandOutput), hvinfoStruct)
			Expect(err).ToNot(HaveOccurred())

			By("checking hvinfo status")
			Expect(hvinfoStruct.HyperVstatus).To(BeTrue())

		})
	})

})
