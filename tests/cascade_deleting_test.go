package tests_test

import (
	"flag"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
	"strings"
)

var _ = Describe("[rfe_id:896][crit:medium][vendor:cnv-qe@redhat.com][level:component]Creation DataVolume via VM yaml file and cascade deleting", func() {

	flag.Parse()

	const (
		featureDv    = "DataVolumes"
		featureQuery = "-o=jsonpath='{.data.feature-gates}'"
		phaseQuery   = "-o=jsonpath='{.status.phase}'"

		dvName        = "cascade-dv"
		cirrosHttpUrl = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
	)

	// The list of resources for check
	var resourceTypes = [3]string{"vm", "dv", "pvc"}

	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	Describe("Verify that DataVolumes feature is enabled by default", func() {
		It("[test_id:835]Verifies the configmap for DataVolumes feature", func() {
			res, _, err := ktests.RunCommandWithNS("kubevirt", "oc", "get", "cm", "kubevirt-config", featureQuery)
			Expect(err).ToNot(HaveOccurred(), "Command should run without errors")
			Expect(strings.Contains(res, featureDv)).To(BeTrue(), "Response should contain DataVolumes feature")
		})
	})

	table.DescribeTable("Create VM with a DataVolume and verify cascade deleting", func(cascadeOption bool) {
		By("Creating new VM with DataVolume")
		dataVolume := ktests.NewRandomDataVolumeWithHttpImport(cirrosHttpUrl, tests.NamespaceTestDefault)
		dataVolume.Name = dvName
		storageClassName := "glusterfs-storage"
		dataVolume.Spec.PVC.StorageClassName = &storageClassName

		vmi := ktests.NewRandomVMIWithDataVolume(dataVolume.Name)
		vmi.Name = dvName
		vm := ktests.NewRandomVirtualMachine(vmi, false)
		vm.Spec.DataVolumeTemplates = append(vm.Spec.DataVolumeTemplates, *dataVolume)

		vm, err = virtClient.VirtualMachine(tests.NamespaceTestDefault).Create(vm)
		Expect(err).ToNot(HaveOccurred())

		By("Verify that VM and all necessary resources were created")
		for _, resource := range resourceTypes {
			obj, err := tests.GetObjects(ktests.NamespaceTestDefault, resource)
			Expect(err).ToNot(HaveOccurred())
			objName := strings.Join(obj[:], " ")
			Expect(strings.Contains(objName, "cascade")).To(BeTrue(), "Resource with name 'cascade' should exists in "+resource)
		}

		By("Waiting until 'importer' pod will be completed and removed")
		tests.WaitUntilResourceDeleted("pod", "importer-cascade")

		By("Starting VM: VMI should appears and get status 'Running'")
		vm = ktests.StartVirtualMachine(vm)
		tests.WaitUntilResourceReadyByNameTestNamespace("vmi", dvName, phaseQuery, "Running")

		if cascadeOption {
			By("Deleting VM with option '--cascade=true'")
			_, _, err = ktests.RunCommand("oc", "delete", "vm", dvName)
			Expect(err).ToNot(HaveOccurred(), "")
		} else {
			By("Deleting VM with option '--cascade=false'")
			_, _, err = ktests.RunCommand("oc", "delete", "vm", dvName, "--cascade=false")
			Expect(err).ToNot(HaveOccurred(), "")

			By("Verifying that VM deleted but all resources still exists")
			for _, resource := range resourceTypes {
				if resource == "vm" {
					tests.WaitUntilResourceDeleted(resource, "cascade")
				} else {
					obj, err := tests.GetObjects(tests.NamespaceTestDefault, resource)
					Expect(err).ToNot(HaveOccurred())
					objName := strings.Join(obj[:], " ")
					Expect(strings.Contains(objName, "cascade")).To(BeTrue(), "Resource with name 'cascade' should exists in "+resource)
				}
			}
			By("Deleting VMI")
			_, _, err = ktests.RunCommand("oc", "delete", "vmi", dvName)
			Expect(err).ToNot(HaveOccurred(), "")

			By("Deleting DataVolume")
			_, _, err = ktests.RunCommand("oc", "delete", "dv", dvName)
			Expect(err).ToNot(HaveOccurred(), "")
		}

		By("Verify that VM and all necessary resources were deleted")
		for _, resource := range resourceTypes {
			tests.WaitUntilResourceDeleted(resource, "cascade")
		}
	},
		table.Entry("[test_id:837] Delete VM, all related resources deleted", true),
		table.Entry("[test_id:838] Delete VM with option '--cascade=false', all related resources still exists", false),
	)
})
