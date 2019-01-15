package tests_test

import (
	"flag"
	"time"

	expect "github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"

	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	"kubevirt.io/kubevirt/pkg/config"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
)

var _ = Describe("[rfe_id:905][crit:medium][vendor:cnv-qe@redhat.com][level:component]Config", func() {

	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	Context("With a non default ServiceAccount created", func() {

		Context("With a single ServiceAccount created and attached", func() {
			var (
				serviceAccountName string
				serviceAccountPath string
			)

			BeforeEach(func() {
				serviceAccountName = "secret-" + uuid.NewRandom().String()
				serviceAccountPath = config.ServiceAccountSourceDir

				tests.CreateServiceAccount(serviceAccountName)
			})

			AfterEach(func() {
				tests.DeleteServiceAccount(serviceAccountName)
			})

			It("[test_id:999]Should be the fs layout the same for a pod and vmi", func() {
				By("Running VMI")
				vmi := ktests.NewRandomVMIWithEphemeralDiskAndUserdataHighMemory(
					ktests.ContainerDiskFor(
						ktests.ContainerDiskFedora), "#!/bin/bash\necho \"fedora\" | passwd fedora --stdin\n")
				ktests.AddServiceAccountDisk(vmi, serviceAccountName)
				ktests.RunVMIAndExpectLaunch(vmi, false, 90)

				By("Checking if ServiceAccount has been attached to the pod")
				vmiPod := ktests.GetRunningPodByVirtualMachineInstance(vmi, tests.NamespaceTestDefault)
				podOutput, err := ktests.ExecuteCommandOnPod(
					virtClient,
					vmiPod,
					vmiPod.Spec.Containers[1].Name,
					[]string{"cat",
						serviceAccountPath + "/namespace",
					},
				)
				Expect(err).To(BeNil())
				Expect(podOutput).To(Equal(tests.NamespaceTestDefault))

				By("Checking mounted serviceaccount image")
				expecter, err := tests.LoggedInFedoraExpecter(vmi.Name, tests.NamespaceTestDefault, 360)
				Expect(err).ToNot(HaveOccurred())
				defer expecter.Close()

				_, err = expecter.ExpectBatch([]expect.Batcher{
					// mount iso Secret image
					&expect.BSnd{S: "sudo su -\n"},
					&expect.BExp{R: "#"},
					&expect.BSnd{S: "mount /dev/sda /mnt\n"},
					&expect.BSnd{S: "echo $?\n"},
					&expect.BExp{R: "0"},
					&expect.BSnd{S: "cat /mnt/namespace\n"},
					&expect.BExp{R: tests.NamespaceTestDefault},
				}, 200*time.Second)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("With two ServiceAccount created and attached", func() {
			var (
				serviceAccountName1 string
				serviceAccountName2 string
			)

			BeforeEach(func() {
				serviceAccountName1 = "secret1-" + uuid.NewRandom().String()
				serviceAccountName2 = "secret2-" + uuid.NewRandom().String()

				tests.CreateServiceAccount(serviceAccountName1)
				tests.CreateServiceAccount(serviceAccountName2)
			})

			AfterEach(func() {
				tests.DeleteServiceAccount(serviceAccountName1)
				tests.DeleteServiceAccount(serviceAccountName2)
			})

			It("[test_id:1001][posneg:negative]Should not run the vm with 2 service accounts", func() {
				By("Ensuring VMI is not running")
				vmi := ktests.NewRandomVMIWithEphemeralDiskAndUserdataHighMemory(
					ktests.ContainerDiskFor(
						ktests.ContainerDiskFedora), "#!/bin/bash\necho \"fedora\" | passwd fedora --stdin\n")
				ktests.AddServiceAccountDisk(vmi, serviceAccountName1)
				ktests.AddServiceAccountDisk(vmi, serviceAccountName2)
				_, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
})
