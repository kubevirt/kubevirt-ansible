package tests_test

import (
	"flag"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
)

var _ = Describe("RBAC", func() {
	flag.Parse()

	var vm tests.VirtualMachine

	const NamespaceTestSystem = "kubevirt-test-system"

	vm.Name = "vm-cirros"
	vm.Manifest = "tests/manifests/vm-cirros.yaml"

	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	BeforeEach(func() {
		By(fmt.Sprintf("Login with the user %s", tests.UsernameAdminUser))
		_, _, err := ktests.RunCommandWithNS("", "oc", "login", "-u", tests.UsernameAdminUser, "-p", "123456")
		Expect(err).ToNot(HaveOccurred(), "Should login as %s user", tests.UsernameAdminUser)

		ktests.BeforeTestCleanup()

		By("Creating a user")
		err = tests.CreateUserWithParameter(tests.UsernameTestUser)
		Expect(err).ToNot(HaveOccurred(), "Creating user %s should not fail: %v", tests.UsernameTestUser, err)
	})

	AfterEach(func() {
		By(fmt.Sprintf("Login with the user %s", tests.UsernameAdminUser))
		_, _, err := ktests.RunCommandWithNS("", "oc", "login", "-u", tests.UsernameAdminUser, "-p", "123456")
		Expect(err).ToNot(HaveOccurred(), "Should login as %s user", tests.UsernameAdminUser)

		By("Deleting a user")
		err = tests.DeleteUserWithParameter(tests.UsernameTestUser)
		Expect(err).ToNot(HaveOccurred(), "Deleting user %s should not fail: %v", tests.UsernameTestUser, err)

		By(fmt.Sprintf("Deleting the identity associated with %s user", tests.UsernameTestUser))
		_, _, err = ktests.RunCommandWithNS("", "oc", "delete", "identity", "allow_all_auth:"+tests.UsernameTestUser)
		Expect(err).ToNot(HaveOccurred(), "Should delete identity of the %s user", tests.UsernameTestUser)
	})

	Describe("RBAC with default permission", func() {
		It("should allow to access subresource endpoint.", func() {
			By(fmt.Sprintf("Login with the user %s", tests.UsernameTestUser))
			_, _, err := ktests.RunCommandWithNS("", "oc", "login", "-u", tests.UsernameTestUser, "-p", "123456")
			Expect(err).ToNot(HaveOccurred(), "Should login as %s user", tests.UsernameTestUser)

			By("Creating a project/namespace")
			err = tests.CreateNamespaceWithParameter(NamespaceTestSystem)
			Expect(err).ToNot(HaveOccurred(), "Should create %s project/namespace: %v", NamespaceTestSystem, err)

			By("Creating a VM from manifest")
			_, _, err = ktests.RunCommandWithNS(NamespaceTestSystem, "oc", "create", "-f", vm.Manifest)
			Expect(err).ToNot(HaveOccurred(), "Should create the VM %q in %s namespace", vm.Manifest, NamespaceTestSystem)

			By("Starting the VM")
			_, _, err = ktests.RunCommandWithNS(NamespaceTestSystem, "virtctl", "start", vm.Name)
			Expect(err).ToNot(HaveOccurred(), "Should schedule the VM %q in %s namespace to start", vm.Name, NamespaceTestSystem)

			By("Checking if the VMI is running")
			Eventually(func() string {
				output, _, err := ktests.RunCommandWithNS(NamespaceTestSystem, "oc", "get", "vmi", vm.Name, "--template", "{{.status.phase}}")
				ExpectWithOffset(1, err).ToNot(HaveOccurred(), "Should get the phase of the VMI %q in %s namespace: output %q: %v", vm.Name, NamespaceTestSystem, output, err)
				return output
			}, time.Minute*10, time.Second*1).Should(Equal("Running"), "The VMI %q in %s namespace should reach \"Running\" phase within 10 minutes", vm.Name, NamespaceTestSystem)

			By("Checking if the VMI's console is accessible")
			vmi, err := virtClient.VirtualMachineInstance(NamespaceTestSystem).Get(vm.Name, &metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred(), "Should be able to fetch the object of the VMI %q in %s namespace", vm.Name, NamespaceTestSystem)
			expecter, err := ktests.LoggedInCirrosExpecter(vmi)
			Expect(err).ToNot(HaveOccurred(), "Should be able to access the console of the VMI %q in %s namespace", vm.Name, NamespaceTestSystem)
			expecter.Close()

			By("Checking if the VMI's VNC server gives the valid response")
			response, err := tests.VNCConnection(NamespaceTestSystem, vm.Name)
			Expect(err).ToNot(HaveOccurred(), "Should open VNC connection to VMI %q in %s namespace", vm.Name, NamespaceTestSystem)
			Expect(response).To(Equal("RFB 003.008"), "Should receive valid response from the VNC connection to the VMI %q in %s namespace", vm.Name, NamespaceTestSystem)

			By("Stopping the VM")
			_, _, err = ktests.RunCommandWithNS(NamespaceTestSystem, "virtctl", "stop", vm.Name)
			Expect(err).ToNot(HaveOccurred(), "VM %q in %s namespace should be scheduled to stop: %v", vm.Name, NamespaceTestSystem, err)

			By("Deleting the VM specified by manifest")
			_, _, err = ktests.RunCommandWithNS(NamespaceTestSystem, "oc", "delete", "-f", vm.Manifest)
			Expect(err).ToNot(HaveOccurred(), "VM %q in %s namespace should be deleted", vm.Name, NamespaceTestSystem)

			By("Deleting the namespace")
			err = tests.RemoveNamespaceWithParameter(NamespaceTestSystem)
			Expect(err).ToNot(HaveOccurred(), "Namespace %s should be deleted: %v", NamespaceTestSystem, err)
		})
	})

	DescribeTable("RBAC Authorization", func(role, namespace string, con bool) {
		By(fmt.Sprintf("Add %s to user %s", role, tests.UsernameTestUser))
		_, _, err := ktests.RunCommandWithNS(namespace, "oc", "adm", "policy", "add-role-to-user", role, tests.UsernameTestUser)
		Expect(err).ToNot(HaveOccurred(), "Role %s should be added to %s user", role, tests.UsernameTestUser)

		if con == false {
			createResourcesToTestViewRole()
		}

		By(fmt.Sprintf("Login with the user %s", tests.UsernameTestUser))
		_, _, err = ktests.RunCommandWithNS("", "oc", "login", "-u", tests.UsernameTestUser, "-p", "123456")
		Expect(err).ToNot(HaveOccurred(), "Should login as %s user", tests.UsernameTestUser)

		By("Creating a VM from manifest")
		_, _, err = ktests.RunCommandWithNS(namespace, "oc", "create", "-f", vm.Manifest)
		if con == true {
			Expect(err).ToNot(HaveOccurred(), "Should schedule the VM %q in %s namespace to start", vm.Name, namespace)
		} else {
			Expect(err).To(HaveOccurred(), "Should not create the VM %q in %s namespace", vm.Manifest, namespace)
		}

		By("Starting the VM")
		_, _, err = ktests.RunCommandWithNS(namespace, "virtctl", "start", vm.Name)
		if con == true {
			Expect(err).ToNot(HaveOccurred(), "Should schedule the VM %q in %s namespace to start", vm.Name, namespace)
		} else {
			Expect(err).To(HaveOccurred(), "Should not schedule the VM %q in %s namespace to start", vm.Name, namespace)
		}

		By("Checking if the VMI is running")
		getVMIStatus := func() string {
			output, _, err := ktests.RunCommandWithNS(namespace, "oc", "get", "vmi", vm.Name, "--template", "{{.status.phase}}")
			if con == true {
				ExpectWithOffset(1, err).ToNot(HaveOccurred(), "Should get the phase of the VMI %q in %s namespace: output %s: %v", vm.Name, namespace, output, err)
			} else {
				ExpectWithOffset(1, err).To(HaveOccurred(), "The phase of the VMI %q in %s namespace shouldn't be accessible: output %s", vm.Name, namespace, output)
			}
			return output
		}

		if con == true {
			Eventually(getVMIStatus, 10*time.Minute, 1*time.Second).Should(Equal("Running"), "The VMI %q in %s namespace should reach the \"Running\" phase within 10 minutes", vm.Name, namespace)
		}

		By("Checking if the VMI's console is accessible")
		vmi, err := virtClient.VirtualMachineInstance(namespace).Get(vm.Name, &metav1.GetOptions{})
		if con == true {
			expecter, err := ktests.LoggedInCirrosExpecter(vmi)
			Expect(err).ToNot(HaveOccurred(), "Should be able to login to the console of the VMI %q in %s namespace", vm.Name, namespace)
			expecter.Close()
		} else {
			_, err = ktests.LoggedInCirrosExpecter(vmi)
			Expect(err).To(HaveOccurred(), "Should not login to the console of the VMI %q in %s namespace", vm.Name, namespace)
		}

		By("Checking if the VMI's VNC server gives the valid response")
		response, err := tests.VNCConnection(namespace, vm.Name)
		if con == true {
			Expect(err).ToNot(HaveOccurred(), "Should open VNC connection to VMI %q in %s namespace", vm.Name, namespace)
			Expect(response).To(Equal("RFB 003.008"), "Should receive a valid VNC connection response from the VMI %q in %s namespace", vm.Name, namespace)
		} else {
			Expect(err).To(HaveOccurred(), "Should not open VNC connection to VMI %q in %s namespace", vm.Name, namespace)
			Expect(response).ToNot(Equal("RFB 003.008"), "Should not receive a valid response from the VNC connection to the VMI %q in %s namespace", vm.Name, namespace)
		}

		By("Check if we can stop VM ")
		_, _, err = ktests.RunCommandWithNS(namespace, "virtctl", "stop", vm.Name)
		if con == true {
			Expect(err).ToNot(HaveOccurred(), "The VM %q in %s namespace should be scheduled to stop: %v", vm.Name, namespace, err)
		} else {
			Expect(err).To(HaveOccurred(), "The VM %q in %s namespace should not be scheduled to stop", vm.Name, namespace)
		}

		By("Deleting the VM specified by manifest")
		_, _, err = ktests.RunCommandWithNS(namespace, "oc", "delete", "-f", vm.Manifest)
		if con == true {
			Expect(err).ToNot(HaveOccurred(), "The VM %q in %s namespace should be deleted", vm.Name, namespace)
		} else {
			Expect(err).To(HaveOccurred(), "The VM %q in %s namespace should not be deleted", vm.Name, namespace)
		}
	},
		Entry("with admin permission should allow to access subresource endpoint", "admin", ktests.NamespaceTestDefault, true),
		Entry("with edit permission should allow to access subresource endpoint", "edit", ktests.NamespaceTestDefault, true),
		Entry("with view permission should not allow to access subresource endpoint", "view", ktests.NamespaceTestAlternative, false))
})

func createResourcesToTestViewRole() {
	// function will setup the enviroment to test view role.
	var vm tests.VirtualMachine
	vm.Name = "vm-cirros"
	vm.Manifest = "tests/manifests/vm-cirros.yaml"

	_, _, err := ktests.RunCommandWithNS("", "oc", "login", "-u", tests.UsernameAdminUser, "-p", "123456")
	Expect(err).ToNot(HaveOccurred(), "Should login as %s user", tests.UsernameAdminUser)

	_, _, err = ktests.RunCommandWithNS(ktests.NamespaceTestAlternative, "oc", "create", "-f", vm.Manifest)
	Expect(err).ToNot(HaveOccurred(), "Should create VM %q in %s namespace", vm.Name, ktests.NamespaceTestAlternative)

	_, _, err = ktests.RunCommandWithNS(ktests.NamespaceTestAlternative, "virtctl", "start", vm.Name)
	Expect(err).ToNot(HaveOccurred(), "Should schedule VM %q in %s namespace to start", vm.Name, ktests.NamespaceTestAlternative)

	Eventually(func() string {
		output, _, err := ktests.RunCommandWithNS(ktests.NamespaceTestAlternative, "oc", "get", "vmi", vm.Name, "--template", "{{.status.phase}}")
		Expect(err).ToNot(HaveOccurred(), "Should get the phase of the VMI %q in %s namespace: output %s: %v", vm.Name, ktests.NamespaceTestAlternative, output, err)
		return output
	}, time.Minute*10, time.Second*1).Should(Equal("Running"), "The VMI %q in %s namespace should reach the \"Running\" phase within 10 minutes", vm.Name, ktests.NamespaceTestAlternative)
}
