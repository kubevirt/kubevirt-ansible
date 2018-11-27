package tests_test

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	ktests "kubevirt.io/kubevirt/tests"
)

var _ = Describe("CDI Upload", func() {
	var pvcName, pvcName1, pvcSize string
	const (
		downloadImgPath     = "/tmp/test.img"
		cdiNamespace        = "cdi"
		uploadRouteFilePath = "tests/manifests/upload-route.yml"
	)
	ktests.BeforeAll(func() {
		By(fmt.Sprintf("downloading cirros Image..."))
		tests.DownloadFile(downloadImgPath, cirrosImgURL)
	})

	Context("virtctl upload using openshift route", func() {
		BeforeEach(func() {
			pvcName = "vm-upload-test-route"
			pvcSize = "500Mi"
		})

		AfterEach(func() {
			By(fmt.Sprintf("clean up route cdi-uploadproxy"))
			args := []string{"delete", "route", "cdi-uploadproxy"}
			_, _, err := ktests.RunCommandWithNS(cdiNamespace, "oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By(fmt.Sprintf("clean up upload pvc"))
			args = []string{"delete", "pvc", pvcName, "--ignore-not-found=true"}
			_, _, err = ktests.RunCommandWithNS(cdiNamespace, "oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By(fmt.Sprintf("clean up upload pod if it still exist"))
			args = []string{"delete", "pod", "cdi-upload-vm-upload-test", "--ignore-not-found=true"}
			_, _, err = ktests.RunCommandWithNS(cdiNamespace, "oc", args...)
			Expect(err).ToNot(HaveOccurred())
		})

		It("will be successful, create again will prompt already exist", func() {
			By(fmt.Sprintf("oc create route cdi-uploadproxy"))
			args := []string{"apply", "-f", uploadRouteFilePath}
			_, _, err := ktests.RunCommandWithNS(cdiNamespace, "oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By(fmt.Sprintf("oc get route cdi-uploadproxy"))
			args = []string{"get", "route", "cdi-uploadproxy", "-o=jsonpath='{.status.ingress[0].host}'"}
			hostURL, _, err := ktests.RunCommandWithNS(cdiNamespace, "oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By(fmt.Sprintf("virtctl upload images via route"))
			args = []string{"image-upload",
				"--image-path=" + downloadImgPath,
				"--pvc-name=" + pvcName,
				"--pvc-size=500Mi",
				"--uploadproxy-url=https://" + hostURL[1:len(hostURL)-1],
				"--insecure"}
			_, _, err = ktests.RunCommandWithNS(cdiNamespace, "virtctl", args...)
			Expect(err).ToNot(HaveOccurred())
			By(fmt.Sprintf("virtctl upload again will be failed"))
			_, _, err = ktests.RunCommandWithNS(cdiNamespace, "virtctl", args...)
			Expect(err).To(HaveOccurred())

		})
	})

	Context("virtctl upload using kubectl port-forward", func() {
		var cmd *exec.Cmd
		BeforeEach(func() {
			pvcName = "vm-upload-test-port-forward"
			pvcName1 = "vm-upload-test-port-forward1"
			pvcSize = "500Mi"
			// Start a process:
			cmd = exec.Command("kubectl", "port-forward", "-n", cdiNamespace, "svc/cdi-uploadproxy", "18443:443")
			if err := cmd.Start(); err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
		})

		AfterEach(func() {
			// Kill it:
			if err := cmd.Process.Kill(); err != nil {
				Expect(err).ToNot(HaveOccurred())
			}

			By(fmt.Sprintf("clean up upload pvc"))
			args := []string{"delete", "pvc", pvcName, "--ignore-not-found=true"}
			_, _, err := ktests.RunCommandWithNS(cdiNamespace, "oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By(fmt.Sprintf("clean up upload pod if it still exist"))
			args = []string{"delete", "pod", "cdi-upload-vm-upload-test", "--ignore-not-found=true"}
			_, _, err = ktests.RunCommandWithNS(cdiNamespace, "oc", args...)
			Expect(err).ToNot(HaveOccurred())
		})

		It("will be successful, create again will prompt already exist", func() {

			By(fmt.Sprintf("virtctl upload images via port-forward"))
			args := []string{"image-upload",
				"--image-path=" + downloadImgPath,
				"--pvc-name=" + pvcName,
				"--pvc-size=" + pvcSize,
				"--uploadproxy-url=https://localhost:18443",
				"--insecure"}
			_, _, err := ktests.RunCommandWithNS(cdiNamespace, "virtctl", args...)
			Expect(err).ToNot(HaveOccurred())

			By(fmt.Sprintf("virtctl upload again will be failed"))
			_, _, err = ktests.RunCommandWithNS(cdiNamespace, "virtctl", args...)
			Expect(err).To(HaveOccurred())
		})

		It("will be failed with invalid uplaod url", func() {

			By(fmt.Sprintf("virtctl upload images with invalid url"))
			args := []string{"image-upload",
				"--image-path=" + downloadImgPath,
				"--pvc-name=" + pvcName1,
				"--pvc-size=" + pvcSize,
				"--uploadproxy-url=https://invalid",
				"--insecure"}
			_, _, err := ktests.RunCommandWithNS(cdiNamespace, "virtctl", args...)
			Expect(err).To(HaveOccurred())
		})

		//clean up image at the end
		By(fmt.Sprintf("clean up cirros Image..."))
		tests.DeleteFile(downloadImgPath)
	})

})
