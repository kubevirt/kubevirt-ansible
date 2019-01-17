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
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"kubevirt.io/kubevirt/tests"
	"os/exec"
)

func search_for_pattern(yaml_segment string, pattern string) bool {
	r, _ := regexp.Compile(pattern)
	for _, line := range strings.Split(yaml_segment, "\n") {
		if r.MatchString(line) {
			return true
		}
	}
	return false
}

var _ = Describe("[rfe_id:1235][crit:medium][vendor:cnv-qe@redhat.com][level:component]Common templates", func() {
	flag.Parse()

	BeforeEach(func() {
		tests.BeforeTestCleanup()
	})

	Context("Testing generated templates", func() {


		It("[test_id:1069]Check if template valid for UI", func() {

			// Getting common_templates
			// TODO: replace downloading common-templates with getting common-templates from RPM

			ct_yml_url_byte, err := exec.Command("/bin/bash", "-c", "curl -s https://api.github.com/repos/kubevirt/common-templates/releases/latest | grep browser_download_url | cut -d '\"' -f 4").Output()
			ct_yml_url := string(ct_yml_url_byte)
			Expect(err).ToNot(HaveOccurred())
			response, err := http.Get(ct_yml_url)
			Expect(err).NotTo(HaveOccurred())
			defer response.Body.Close()
			data, err := ioutil.ReadAll(response.Body)
			common_templates_yaml := string(data)

			for _, yaml_segment := range strings.Split(common_templates_yaml, "\n---\n") {

				if !search_for_pattern(yaml_segment, "^Kind: Template$") {
					continue
				}

				By("Checking that template contains required lables")
				Expect(search_for_pattern(yaml_segment, "\\s+os.template.cnv.io/[a-z0-9\\.]+:\\s\"true\"$")).To(BeTrue(), "Template should have os label")
				Expect(search_for_pattern(yaml_segment, "\\s+workload.template.cnv.io/[a-z]+:\\s\"true\"$")).To(BeTrue(), "Template should have workload label")
				Expect(search_for_pattern(yaml_segment, "\\s+flavor.template.cnv.io/[a-z]+:\\s\"true\"$")).To(BeTrue(), "Template should have flavor label")
				Expect(search_for_pattern(yaml_segment, "\\s+template.cnv.io/type:\\s\"base\"$")).To(BeTrue(), "Template should have type base")

			}
		})
	})
})
