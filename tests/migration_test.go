package tests_test

import (
	"flag"
	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

		cpuModel = "Nehalem"

		migrationMethodLive  = "LiveMigration"
		migrationMethodBlock = "BlockMigration"

		sharedPVC    = "ReadWriteMany"
		nonsharedPVC = "ReadWriteOnce"

		phaseQuery           = "-o=jsonpath='{.status.phase}'"
		nodeQuery            = "-o=jsonpath='{.status.nodeName}'"
		migrationMethodQuery = "-o=jsonpath='{.status.migrationMethod}'"

		PvcCdiFilePath = "tests/manifests/golden-pvc.yml"
		VmCdiFilePath  = "tests/manifests/test-vm.yml"
		dstPVCFilePath = "/tmp/test-pvc.json"
		dstVMFilePath  = "/tmp/test-vm.json"

		migrationJobTemplate = "tests/manifests/migration-job.yml"
		temporaryJson        = "/tmp/tmp-job.yml"

		cirrosHttpLink  = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
		cirrosImageLink = "kubevirt/cirros-container-disk-demo:latest"

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
			expecter, err = tests.LoggedInCirrosExpecter(vmName, tests.NamespaceTestDefault, 180)
			Expect(err).ToNot(HaveOccurred(), "Should be able to access the console of the VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)
		case "fedora":
			expecter, err = tests.LoggedInFedoraExpecter(vmName, tests.NamespaceTestDefault, 360)
			Expect(err).ToNot(HaveOccurred(), "Should be able to access the console of the VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)
		}

		rndFile := "testFile-" + rand.String(5)
		expecter.Send("touch " + rndFile + "\n")
		expecter.Expect(regexp.MustCompile(`\$`), 5*time.Second)
		//expecter.Send("\x1d")
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

		}, tests.ShortTimeout, 1*time.Second).Should(Equal(true))

		By("Get Node where VMI is located after migration")
		targetNode := tests.GetResourceSpecificParameters("vmi", vmName, nodeQuery)
		Expect(targetNode).NotTo(Equal(sourceNode), "VM should be located on another Node. Source node: %sn, Target node: %tn", sourceNode, targetNode)

		By("Check if the VMI's console is accessible after migration")
		expecter, _, err = ktests.NewConsoleExpecter(virtClient, vmi, 10*time.Second)
		Expect(err).ToNot(HaveOccurred(), "Should be able to access the console of the VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)
		_, err = expecter.ExpectBatch([]expect.Batcher{
			&expect.BSnd{S: "\n"},
			&expect.BSnd{S: "\n"},
			&expect.BExp{R: "\\$ "},
			&expect.BSnd{S: "ls -la\n"},
			&expect.BExp{R: rndFile},
			&expect.BSnd{S: "rm " + rndFile + "\n"},
			&expect.BExp{R: "\\$ "},
		}, 60*time.Second)
		Expect(err).ToNot(HaveOccurred())
		//expecter.Send("\x1d")
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

		It("Verifies VM with shared PVC CDI", func() {
			vmName := "vm-pvc-cdi"

			tests.ProcessTemplateWithParameters(PvcCdiFilePath, dstPVCFilePath, "PVC_NAME="+vmName, "EP_URL="+cirrosHttpLink, "ACCESS_MODE="+sharedPVC)
			tests.CreateResourceWithFilePathTestNamespace(dstPVCFilePath)
			tests.WaitUntilResourceReadyByNameTestNamespace("pvc", vmName, "-o=jsonpath='{.metadata.annotations}'", "pv.kubernetes.io/bind-completed:yes")
			tests.WaitUntilResourceDeleted("pod", "importer-"+vmName)

			tests.ProcessTemplateWithParameters(VmCdiFilePath, dstVMFilePath, "VM_NAME="+vmName, "PVC_NAME="+vmName, "VM_APIVERSION="+vmAPIVersion)
			tests.CreateResourceWithFilePathTestNamespace(dstVMFilePath)

			runMigrationAndExpectCompletion(vmName, migrationMethodLive, osCirros)
		})

		It("Verifies VM with containerDisk + shared PVC", func() {
			PVCName := "container-pvc"
			storageClassName := "glusterfs-storage"

			vmi := ktests.NewRandomVMIWithEphemeralDisk(ktests.ContainerDiskFor(ktests.ContainerDiskCirros))
			tests.AddСPU(vmi, 1, cpuModel)
			tests.CreatePVC(PVCName, "1Gi", storageClassName, k8sv1.ReadWriteMany)
			ktests.AddPVCDisk(vmi, PVCName, "virtio", PVCName)
			vmi, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
			Expect(err).To(BeNil())

			runMigrationAndExpectCompletion(vmi.Name, migrationMethodBlock, osCirros)
		})

		It("VM with NFS shared PVC + CloudInit + ServiceAccount", func() {
			PVName := "nfs-pvc"
			path := "/PVS/dshchedr/fedora"
			server := "10.9.96.20"
			size := "5Gi"

			serviceAccountName := "secret-" + rand.String(5)
			tests.CreateServiceAccount(serviceAccountName)

			tests.CreateNFSPvAndPvc(PVName, size, path, server, k8sv1.ReadWriteMany)
			vmi := ktests.NewRandomVMIWithPVC(PVName)
			vmi.Spec.Domain.Resources.Requests[k8sv1.ResourceMemory] = resource.MustParse("2G")
			tests.AddСPU(vmi, 1, cpuModel)
			ktests.AddUserData(vmi, "cloud-init", "#cloud-config\npassword: fedora\nchpasswd: { expire: False }\nbootcmd:\n- \"mount /dev/sda /mnt\"")
			ktests.AddServiceAccountDisk(vmi, serviceAccountName)

			vmi, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
			Expect(err).To(BeNil())

			runMigrationAndExpectCompletion(vmi.Name, migrationMethodBlock, osFedora)

			By("Checking that serviceAccount still mounted after migration")
			expecter, _, err := ktests.NewConsoleExpecter(virtClient, vmi, 10*time.Second)
			Expect(err).ToNot(HaveOccurred(), "Should be able to access the console of the VMI %q in %s namespace", vmName, ktests.NamespaceTestDefault)

			_, err = expecter.ExpectBatch([]expect.Batcher{
				&expect.BSnd{S: "\n"},
				&expect.BExp{R: "\\$ "},
				&expect.BSnd{S: "cat /mnt/namespace\n"},
				&expect.BExp{R: tests.NamespaceTestDefault},
			}, timeout)
			Expect(err).ToNot(HaveOccurred())
			expecter.Close()
		})

	})

	// TO DO

	//// scenarios with bugs
	// "VM with containerDisk + CloudInit + ServiceAccount + ConfigMap + Secret"
	// "VM with DataVolume shared PVC"
	// "Negative: VM with non-shared PVC"

	//// covered by tier1
	// "VM with containerDisk"
	// "VM with iSCSI PVC"

})
