package tests_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"kubevirt.io/kubevirt-ansible/tests"
)

// template parameters
const (
	v2vServiceInstance    = "tests/manifests/import-vm-v2v-apb.yml"
	dstv2vServiceInstance = "/tmp/test-v2v.json"
)

// Array of enviroment variables to get from testing env
var listofEnvVars = []string{"ocpID", "ocpPass", "VMWareURL", "VMName", "VMWareID", "VMWarePass"}

// Map to store Enviroment variables
var v2vEnvVars map[string]string

// Get all the Env variable needed for tests
func getV2vEnv() {
	v2vEnvVars = map[string]string{}
	for _, vars := range listofEnvVars {
		val, ok := os.LookupEnv(vars)
		// skip the vars that are not found
		if ok {
			v2vEnvVars[vars] = val
		}
	}
}

var _ = Describe("Importing VMs using V2V", func() {

	BeforeEach(func() {
		getV2vEnv()
		if len(listofEnvVars) != len(v2vEnvVars) {
			Skip("Required Enviroment variables not found")
		}
	})

	Context("Create ServiceInstance of import-vm-apb", func() {
		DescribeTable("v2v import with apb Plan",
			func(apbPlanName, verifyStr, verifyStatus string) {
				tests.ProcessTemplateWithParameters(v2vServiceInstance, dstv2vServiceInstance,
					"APB_PLAN_NAME="+apbPlanName,
					"OCP_ID="+v2vEnvVars["ocpID"],
					"OCP_PASS="+v2vEnvVars["ocpPass"],
					"VMWARE_URL="+v2vEnvVars["VMWareURL"],
					"VM_NAME="+v2vEnvVars["VMName"],
					"VMWARE_ID="+v2vEnvVars["VMWareID"],
					"VMWARE_PASS="+v2vEnvVars["VMWarePass"])
				tests.CreateResourceWithFilePathTestNamespace(dstv2vServiceInstance)
				tests.WaitUntilResourceReadyByNameTestNamespace("vm", v2vEnvVars["VMName"], "-o=jsonpath='"+verifyStr+"'", verifyStatus)
			},
			Entry("vmware", "vmware", "{.status.created}", "true"),
			Entry("vmware-template", "vmware-template", "{.metadata.name}", v2vEnvVars["VMName"]),
		)
	})
})
