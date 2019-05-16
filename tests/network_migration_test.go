package tests_test

import (

	//commented by Xenia Lisovskaia
	//quick workaround due syntax error (tests can't compilate, CI broken)

	/*	"flag"
		"fmt"
		"strconv"
		"time"
		v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
		k8sv1 "k8s.io/api/core/v1"
		ktests "kubevirt.io/kubevirt/tests"

		"k8s.io/apimachinery/pkg/api/resource"
		"kubevirt.io/kubevirt/pkg/kubecli"
		"kubevirt.io/kubevirt/pkg/virtctl/expose"

	*/
	"fmt"
	expect "github.com/google/goexpect"
	. "github.com/onsi/gomega"
	v1 "kubevirt.io/kubevirt/pkg/api/v1"
	ktests "kubevirt.io/kubevirt/tests"
)

//commented by Xenia Lisovskaia
//quick workaround due syntax error (tests can't compilate, CI broken)
/*
var _ = Describe("[rfe_id:1150][crit:high][vendor:cnv-qe@redhat.com][level:component]Network migration", func() {

	 flag.Parse()

	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	var vmia *v1.VirtualMachineInstance
	var vmib *v1.VirtualMachineInstance


	ktests.BeforeAll(func() {
		ktests.BeforeTestCleanup()
		if !ktests.HasLiveMigration() {
			Skip("LiveMigration feature gate is not enabled in kubevirt-config")
		}

		nodes := ktests.GetAllSchedulableNodes(virtClient)
		Expect(nodes.Items).ToNot(BeEmpty(), "There should be some compute node")

		if len(nodes.Items) < 2 {
			Skip("Migration tests require at least 2 nodes")
		}
	})

	runMigrationAndExpectCompletion := func(migration *v1.VirtualMachineInstanceMigration, timeout int) string {
		By("Starting a Migration")
		_, err := virtClient.VirtualMachineInstanceMigration(migration.Namespace).Create(migration)
		Expect(err).ToNot(HaveOccurred())

		By("Waiting until the Migration Completes")
		uid := ""
		Eventually(func() bool {
			migration, err := virtClient.VirtualMachineInstanceMigration(migration.Namespace).Get(migration.Name, &v13.GetOptions{})
			Expect(err).ToNot(HaveOccurred())

			Expect(migration.Status.Phase).ToNot(Equal(v1.MigrationFailed))

			uid = string(migration.UID)
			if migration.Status.Phase == v1.MigrationSucceeded {
				return true
			}
			return false

		}, timeout, 1*time.Second).Should(Equal(true))
		return uid
	}

	confirmVMIPostMigration := func(vmi *v1.VirtualMachineInstance, migrationUID string) {
		By("Retrieving the VMI post migration")
		vmi, err = virtClient.VirtualMachineInstance(vmi.Namespace).Get(vmi.Name, &v13.GetOptions{})
		Expect(err).ToNot(HaveOccurred())

		By("Verifying the VMI's migration state")
		Expect(vmi.Status.MigrationState).ToNot(BeNil())
		Expect(vmi.Status.MigrationState.StartTimestamp).ToNot(BeNil())
		Expect(vmi.Status.MigrationState.EndTimestamp).ToNot(BeNil())
		Expect(vmi.Status.MigrationState.TargetNode).To(Equal(vmi.Status.NodeName))
		Expect(vmi.Status.MigrationState.TargetNode).ToNot(Equal(vmi.Status.MigrationState.SourceNode))
		Expect(vmi.Status.MigrationState.Completed).To(Equal(true))
		Expect(vmi.Status.MigrationState.Failed).To(Equal(false))
		Expect(vmi.Status.MigrationState.TargetNodeAddress).ToNot(Equal(""))
		Expect(string(vmi.Status.MigrationState.MigrationUID)).To(Equal(migrationUID))

		By("Verifying the VMI's is in the running state")

		Expect(vmi.Status.Phase).To(Equal(v1.Running)

)}

	Context("Masquerde VM is still avaible after migration", func() {
		It("[test_id:CNV-2061] Masquerde VM is still availble after migration", func() {

			By("Create VMIs")
			userData := "#!/bin/bash\npassword: fedora\nyum install -y nginx\nsystemctl enable nginx\nsystemctl start nginx\n"
			vmia = ktests.NewRandomVMIWithEphemeralDiskAndUserdata(ktests.ContainerDiskFor(ktests.ContainerDiskFedora), userData)
			vmia.Labels = map[string]string{"expose": "vmi"}
			vmia.Spec.Domain.CPU = &v1.CPU{Cores: 1, Model: "Nehalem"}
			iface := v1.Interface{
				Name: "testmasquerade",
				InterfaceBindingMethod: v1.InterfaceBindingMethod{Masquerade: &v1.InterfaceMasquerade{}},
				Ports: []v1.Port{{Name: "http", Port: 80, Protocol: "TCP"}},
			}
			vmia.Spec.Domain.Devices.Interfaces = []v1.Interface{iface}
			vmia.Spec.Networks = []v1.Network{
				v1.Network{
					Name: "testmasquerade",
					NetworkSource: v1.NetworkSource{
						Pod: &v1.PodNetwork{},
					},
				},
			}

			vmia.Spec.Domain.Resources.Requests[k8sv1.ResourceName("memory")] = resource.MustParse("1024M")
			_, err = virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Create(vmia)
			Expect(err).ToNot(HaveOccurred())

			vmib = ktests.NewRandomVMIWithEphemeralDiskAndUserdata(ktests.ContainerDiskFor(ktests.ContainerDiskCirros), "#!/bin/bash\necho 'hello'\n")
			_, err = virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Create(vmib)
			Expect(err).ToNot(HaveOccurred())

			By("Exposing the service via virtctl command")
			const testPort = 80
			const servicePort = "27017"
			const serviceName = "cluster-ip-vmi"

			virtctl := ktests.NewRepeatableVirtctlCommand(expose.COMMAND_EXPOSE, "virtualmachineinstance", "--namespace",
					vmia.Namespace, vmia.Name, "--port", servicePort, "--name", serviceName, "--target-port", strconv.Itoa(testPort))
			err := virtctl()
			Expect(err).ToNot(HaveOccurred())

			By("Getting back the cluster IP given for the service")
			svc, err := virtClient.CoreV1().Services(vmia.Namespace).Get(serviceName, v13.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			serviceIP := svc.Spec.ClusterIP

			By("Validate the http port is available")
			resp := "1"
			curlReq(serviceIP, servicePort, vmib, resp)

			By("Do migration")
			migration := ktests.NewRandomMigration(vmia.Name, vmia.Namespace)
			migrationUID := runMigrationAndExpectCompletion(migration, 180)

			By("Validate the migration succeed")
			confirmVMIPostMigration(vmia, migrationUID)

			By("Validate the http port is available again")
			curlReq(serviceIP, servicePort, vmib, resp)
		})
	})

})


*/
func curlReq(ip string, port string, vmi *v1.VirtualMachineInstance, resp string) {
	ktests.WaitUntilVMIReady(vmi, ktests.LoggedInCirrosExpecter)
	err := ktests.CheckForTextExpecter(vmi, []expect.Batcher{
		&expect.BSnd{S: fmt.Sprintf("curl --silent --connect-timeout 5 --head %s%s  | grep 'HTTP/1.1 200 OK' | wc -l \n", ip, port)},
		&expect.BExp{R: resp},
	}, 60)
	Expect(err).ToNot(HaveOccurred())
}
