package tests_test

import (
	"flag"
	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
	"regexp"
	"strings"
	"time"
)

var _ = Describe("Migration", func() {

	flag.Parse()

	const (
		featureMigration = "LiveMigration"
		featureQuery     = "-o=jsonpath='{.data.feature-gates}'"

		migrationMethodLive  = "LiveMigration"
		migrationMethodBlock = "BlockMigration"

		sharedPVC = "ReadWriteMany"
		nonsharedPVC = "ReadWriteOnce"

		phaseQuery           = "-o=jsonpath='{.status.phase}'"
		nodeQuery            = "-o=jsonpath='{.status.nodeName}'"
		migrationMethodQuery = "-o=jsonpath='{.status.migrationMethod}'"

		//migrationJobTemplate  = "tests/manifests/migrationManifests/migration-job.yml"
		migrationJobTemplate = "manifests/migrationManifests/migration-job.yml"
		temporaryJson        = "/tmp/tmp-vm.yml"

		cirrosHttpLink       = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"

		// Located in e2e_test.go
		//PVCFilePath     = "tests/manifests/golden-pvc.yml"
		//VMFilePath      = "tests/manifests/test-vm.yml"
		PVCFilePath     = "manifests/golden-pvc.yml"
		VMFilePath      = "manifests/test-vm.yml"
		dstPVCFilePath = "/tmp/test-pvc.json"
		dstVMFilePath = "/tmp/test-vm.json"

    	osCirros = "cirros"
		osFedora = "fedora"
	)

	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	runMigrationAndExpectCompletion := func(vmName, migrationMethod, os string) {
		By("Waiting until VMI is running")
		tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vmName, phaseQuery, "Running")

		By("Get current Node where VMI is located")
		sourceNode := tests.GetResourceSpecificParameters("vmi", vmName, nodeQuery)

		By("Get migration method")
		mtd := tests.GetResourceSpecificParameters("vmi", vmName, migrationMethodQuery)
		Expect(strings.Contains(mtd, migrationMethod)).To(BeTrue(), "Should be %s", migrationMethod)

		By("Check if the VMI's console is accessible before migration")
		vmi, err := virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Get(vmName, &metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred(), "Should be able to fetch the object of the VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)

		var expecter expect.Expecter
		switch os {
		case "cirros":
			expecter, err = ktests.LoggedInCirrosExpecter(vmi)
			Expect(err).ToNot(HaveOccurred(), "Should be able to access the console of the VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)
		case "fedora":
			expecter, err = ktests.LoggedInFedoraExpecter(vmi)
			Expect(err).ToNot(HaveOccurred(), "Should be able to access the console of the VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)
		}

		rndFile := "testFile-"+rand.String(5)
		expecter.Send("touch "+rndFile+"\n")
		expecter.Expect(regexp.MustCompile("\\$"), 5*time.Second)
		expecter.Send("\x1d")
		expecter.Close()

		By("Check if the VMI's VNC server gives the valid response before migration")
		response, err := tests.VNCConnection(ktests.NamespaceTestDefault, vmName)
		Expect(err).ToNot(HaveOccurred(), "Should open VNC connection to VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)
		Expect(response).To(Equal("RFB 003.008"), "Should receive valid response from the VNC connection to the VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)

		By("Run VM Migration")
		tests.ProcessTemplateWithParameters(migrationJobTemplate, temporaryJson, "VM_NAME="+vmName)
		_, _, err = ktests.RunCommand("oc", "create", "-f", temporaryJson)
		Expect(err).ToNot(HaveOccurred(), "Migration job should be executed without errors")

		By("Wait until migration completed/failed")
		Eventually(func() bool {
			migration := tests.GetResourceSpecificParameters("virtualmachineinstancemigration.kubevirt.io", "job1", phaseQuery)
			Expect(strings.Contains(migration, "Failed")).To(BeFalse())

			if strings.Contains(migration, "Succeeded") {
				return true
			}
			return false

		}, timeout, 1*time.Second).Should(Equal(true))

		By("Get Node where VMI is located after migration")
		targetNode := tests.GetResourceSpecificParameters("vmi", vmName, nodeQuery)
		Expect(targetNode).NotTo(Equal(sourceNode), "VM should be located on another Node. Source node: %sn, Target node: %tn", sourceNode, targetNode)

		By("Check if the VMI's console is accessible after migration")
		expecter, _, err = ktests.NewConsoleExpecter(virtClient, vmi, 10*time.Second)
		Expect(err).ToNot(HaveOccurred(), "Should be able to access the console of the VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)
		expecter.Send("ls -la")
		expecter.Expect(regexp.MustCompile(rndFile), 5*time.Second)
		expecter.Send("\x1d")
		expecter.Close()

		By("Check if the VMI's VNC server gives the valid response after migration")
		response, err = tests.VNCConnection(ktests.NamespaceTestDefault, vmName)
		Expect(err).ToNot(HaveOccurred(), "Should open VNC connection to VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)
		Expect(response).To(Equal("RFB 003.008"), "Should receive valid response from the VNC connection to the VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)

	}


	Describe("Verify configmap for LiveMigration feature", func() {
		It("Verifies the configmap for LiveMigration feature", func() {
			res, _, err := ktests.RunCommandWithNS("kubevirt", "oc", "get", "cm", "kubevirt-config", featureQuery)
			Expect(err).ToNot(HaveOccurred(), "Command should run without errors")
			Expect(strings.Contains(res, featureMigration)).To(BeTrue(), "Response should contain LiveMigration feature")
		})
	})

	Describe("Verify migration of VM with different types of disks", func() {

		// Covered in kubevirt, tier1
		//It("Verifies VM with containerDisk", func() {
		//	var vm tests.VirtualMachine
		//	vm.Name = "vm-container"
		//	vm.Manifest = "manifests/migrationManifests/vm-container.yml"
		//	vm.Namespace = tests.NamespaceTestDefault
		//
		//	By("Creating a VM from manifest and run it")
		//	_, _, err := vm.Create()
		//	Expect(err).ToNot(HaveOccurred(), "VM 'creating' command should be executed without errors")
		//	_, _, err = vm.Start()
		//	Expect(err).ToNot(HaveOccurred(), "VM should start without errors")
		//
		//	runMigrationAndExpectCompletion(vm.Name, migrationMethodBlock, "cirros")
		//})

		FIt("Verifies VM with containerDisk", func() {
			pvcName := "pvc-cdi"
			vmName := "vm-pvc-cdi"
			tests.ProcessTemplateWithParameters(PVCFilePath, dstPVCFilePath, "PVC_NAME="+pvcName, "EP_URL="+cirrosHttpLink, "ACCESS_MODE="+sharedPVC)
			tests.CreateResourceWithFilePathTestNamespace(dstPVCFilePath)
			tests.WaitUntilResourceReadyByNameTestNamespace("pvc", pvcName, "-o=jsonpath='{.metadata.annotations}'", "pv.kubernetes.io/bind-completed:yes")
			tests.ProcessTemplateWithParameters(VMFilePath, dstVMFilePath, "VM_NAME="+vmName, "PVC_NAME="+pvcName, "VM_APIVERSION="+vmAPIVersion)
			tests.CreateResourceWithFilePathTestNamespace(dstVMFilePath)

			runMigrationAndExpectCompletion(vmName, migrationMethodLive, osCirros)
		})
	})


	table.DescribeTable("Verify LiveMigration", func(vmName, vmManifest, migrationMethod, os string) {
		var vm tests.VirtualMachine
		vm.Name = vmName
		vm.Manifest = vmManifest

		By("Creating a VM from manifest and run it")
		_, _, err := vm.Create()
		Expect(err).ToNot(HaveOccurred(), "VM 'creating' command should be executed without errors")
		_, _, err = vm.Start()
		Expect(err).ToNot(HaveOccurred(), "VM should start without errors")

		runMigrationAndExpectCompletion(vm.Name, migrationMethod, os)

	},
		table.Entry("VM with containerDisk + CloudInit + ServiceAccount", "vm-container-cloud-servacc", "manifests/migrationManifests/vm-containerDisk-cloudInit-servAcc.yml", migrationMethodBlock, osFedora),
		table.Entry("VM with containerDisk + shared PVC", "vm-container-pvc", "manifests/migrationManifests/vm-containerDisk-PVC.yml", migrationMethodBlock, osCirros),
	    table.Entry("VM with shared PVC CDI", "vm-pvc-cdi", "manifests/migrationManifests/vm-pvc-cdi.yml", migrationMethodLive, osCirros),

	    table.Entry("VM with NFS shared PVC + CloudInit + ServiceAccount", "vm", "yml", migrationMethodBlock, osFedora),
	    table.Entry("VM with iSCSI PVC", "vm", "yml", migrationMethodLive, osCirros),
		table.Entry("VM with DataVolume shared PVC", "vm", "yml", migrationMethodLive, osCirros),
		table.Entry("Negative: VM with non-shared PVC", "vm", "yml", migrationMethodBlock, osCirros),
	)

})
