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
 * Copyright 2018 Red Hat, Inc.
 *
 */

package tests_test

import (
	"flag"
	"os/exec"
	"strings"
	"net/http"
	"io/ioutil"
	"regexp"
	"os"
	"bufio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/tests"
)

var _ = Describe("Common templates", func() {
	flag.Parse()

	virtClient, err := kubecli.GetKubevirtClient()
	tests.PanicOnError(err)

	// Common templates read into string
	var common_templates_yaml string

	BeforeEach(func() {
		tests.BeforeTestCleanup()

		// Setting namespace
		_, err := exec.Command("/bin/bash", "-c", "/usr/bin/oc project " + tests.NamespaceTestDefault).Output()
		tests.PanicOnError(err)

		// Applying datavolume
		// TODO1: This is currently bugged and replaced with using vm.yaml. 
		// exec.Command("/bin/bash", "-c", "/usr/bin/oc apply -f tests/manifests/fedora_datavolume.yaml")

		// Getting common_templates
		// TODO2: replace downloading common-templates with getting common-templates from RPM
		ct_yml_url := "https://github.com/kubevirt/common-templates/releases/download/v0.3.1/common-templates-v0.3.1.yaml"
		response, err := http.Get(ct_yml_url)
		tests.PanicOnError(err)
		defer response.Body.Close()
		data, err := ioutil.ReadAll(response.Body)
		common_templates_yaml = string(data)

	})

	Context("Test loading Fedora", func() {

		// CNV-1065
		It("Test loading Fedora", func() {

			// Patterns needed for parsing common-templates.yaml
			source_pattern, _ := regexp.Compile("# Source: dist/templates/(.*\\.yaml)\\s*")
			os_pattern, _ := regexp.Compile("([^-]+)-[^-]+-[^-]+\\.yaml")

			for _, yaml_segment := range strings.Split(common_templates_yaml, "\n---\n") {

				lines := strings.Split(yaml_segment, "\n")
				// Skip this yaml segment if it doesn't contain VM template
				if (len(lines) == 0 || ! source_pattern.MatchString(lines[0])) {
					continue
				}

				filename := source_pattern.FindStringSubmatch(lines[0])[1]
				os_name := os_pattern.FindStringSubmatch(filename)[1]
				// Skip this template if it is not Fedora
				if (os_name != "fedora") {
					continue
				}

				By("Reading " + filename)
				// Write data to a file and feed this file to oc process
				// TODO2: replace downloading common-templates with getting common-templates from RPM
				// Currently this core is commented and the workaround is used
				//out, err := os.Create("_out/" + filename)
				//Expect(err).ToNot(HaveOccurred())
				//out.WriteString(yaml_segment)
				//out.Close()
				//volume_name := "fedora-dv"
				// command := "/usr/bin/oc process --local"
				// command += " -f _out/" + filename
				// command += " PVCNAME=" + volume_name
				// command += " NAME=common-templates-test-vm"
				// command += " | /usr/bin/oc apply -f -"
				//output, _ := exec.Command("/bin/bash", "-c", ).Output()

				// The workaround
				_, err := exec.Command("/bin/bash", "-c", "/usr/bin/oc apply -f tests/manifests/vm-fedora-workaround.yaml").Output()
				Expect(err).ToNot(HaveOccurred())

				By("Getting VM")
				getVMOptions := metav1.GetOptions{}
				vm, err := virtClient.VirtualMachine(tests.NamespaceTestDefault).Get("common-templates-test-vm", &getVMOptions)
				Expect(err).ToNot(HaveOccurred())

				By("Starting VM and checking that instance is created")
				vm = tests.StartVirtualMachine(vm)
				getOptions := metav1.GetOptions{}
				_, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Get("common-templates-test-vm", &getOptions)
				Expect(err).ToNot(HaveOccurred())

				By("Checking that pod was created and has the right name")
				listOptions := metav1.ListOptions{}
				podList, err := virtClient.CoreV1().Pods(tests.NamespaceTestDefault).List(listOptions)
				Expect(err).ToNot(HaveOccurred())
				Expect(podList.Items).To(HaveLen(1), "Exactly 1 pod should exist")
				Expect(podList.Items[0].Name).To(HavePrefix("virt-launcher-common-templates-test-vm"), "Pod's name should contain name of the VM associated with it")

				By("Checking output of VMI console")
				cmd := exec.Command("/usr/bin/virtctl", "console", "common-templates-test-vm")
				stdout, err := cmd.StdoutPipe()
				Expect(err).ToNot(HaveOccurred())
				// TODO3: passing stdin as an input because virtctl console require input object to be terminal
				// This later leads to problem with output, but at the moment I don't know how to avoid it
				cmd.Stdin = os.Stdin
				err = cmd.Start()
				Expect(err).ToNot(HaveOccurred())
				scanner := bufio.NewScanner(stdout)
				// We consider VM started successfully if we see these 2 lines one after another
				os_version_pattern, _ := regexp.Compile("^Fedora \\d{2} \\(Cloud Edition\\)")
				kernel_version, _ := regexp.Compile("^Kernel [a-z\\d\\.\\-_]+ on an x86_64 \\(ttyS0\\)")
				os_welcome_found := false
				for scanner.Scan() {
					text := scanner.Text()
					if (os_version_pattern.MatchString(text)) {
						scanner.Scan()
						text = scanner.Text()
						if (kernel_version.MatchString(text)) {
							os_welcome_found = true
							break
						}
					}
				}
				Expect(os_welcome_found).To(BeTrue(), "VM console booting output should reach welcoming lines")
				//err = cmd.Process.Kill()
				//Expect(err).ToNot(HaveOccurred())
				//err = cmd.Wait()
				//Expect(err).ToNot(HaveOccurred())
			}
		})
	})
})

