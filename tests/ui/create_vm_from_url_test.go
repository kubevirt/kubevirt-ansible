package ui_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmcvetta/randutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

var _ = Describe("Create VM from URL via Create With Wizard", func() {
	var page *agouti.Page
	var rstring string
	var vm string

	BeforeEach(func() {
		var err error
		rstring, err = randutil.String(8, randutil.Alphanumeric)
		Expect(err).NotTo(HaveOccurred())
		rstring = strings.ToLower(rstring)

		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
	})

	It("Create virtual machine from URL", func() {
		By("Launch the console page", func() {
			Expect(page.Navigate(url)).To(Succeed())
		})
		By("Fill username and password", func() {
			Eventually(page.FindByName("username"), timeout, pollInterval).Should(BeFound())
			Expect(page.FindByName("username").Fill(username)).To(Succeed())
			time.Sleep(1 * time.Second)
			Expect(page.FindByID("inputPassword").Fill(password)).To(Succeed())
		})
		By("Click button login to login the console", func() {
			Expect(page.FindByButton("Log In").Submit()).To(Succeed())
		})
		By(fmt.Sprintf("Use the project %s for test", uinamespace), func() {
			Eventually(page.FindByClass("co-namespace-selector"), timeout, pollInterval).Should(BeFound())
			Expect(page.FindByClass("co-namespace-selector").Click()).To(Succeed())
			time.Sleep(1 * time.Second)
			Expect(page.AllByLink(uinamespace).At(0).Click()).To(Succeed())
		})
		By("Launch the wizard", func() {
			Eventually(page.FindByButton("Create Virtual Machine"), timeout, pollInterval).Should(BeFound())
			Expect(page.FindByButton("Create Virtual Machine").Click()).To(Succeed())
			Expect(page.FindByLink("Create with Wizard").Click()).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Fill virtual machine name", func() {
			vm = "create-vm-from-url-" + rstring
			Expect(page.FindByXPath("/html/body/div[3]/div[2]/div/div/div/div[2]/section/div/div/form/div[1]/div[2]/input").Fill(vm)).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Fill virtual machine description", func() {
			description := fmt.Sprintf("Description of %s", vm)
			Expect(page.FindByXPath("/html/body/div[3]/div[2]/div/div/div/div[2]/section/div/div/form/div[2]/div[2]/textarea").Fill(description)).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		/* Only enable when no project select at the first page
		By("Choose namespace", func() {
			Expect(page.FindByID("namespace-dropdown").Click()).To(Succeed())
			Expect(page.AllByLink(uinamespace).At(-1).Click()).To(Succeed())
		})
		*/
		By("Choose 'URL' as provision source", func() {
			Expect(page.FindByID("image-source-type-dropdown").Click()).To(Succeed())
			Expect(page.FindByLink("URL").Click()).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Fill image URL", func() {
			url := "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
			Expect(page.FindByXPath("/html/body/div[3]/div[2]/div/div/div/div[2]/section/div/div/form/div[4]/div[2]/input").Fill(url)).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Choose Operating System", func() {
			Expect(page.FindByButton("--- Select Operating System ---").Click()).To(Succeed())
			Expect(page.FindByLink("fedora29").Click()).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Choose VM Flavor", func() {
			Expect(page.FindByButton("--- Select Flavor ---").Click()).To(Succeed())
			Expect(page.FindByLink("small").Click()).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Choose VM Workload Profile ", func() {
			Expect(page.FindByButton("--- Select Workload Profile ---").Click()).To(Succeed())
			Expect(page.FindByLink("generic").Click()).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Check start virtual machine on creation", func() {
			Expect(page.FindByLabel("Start virtual machine on creation").Check()).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Click button 'Next' to go to network page", func() {
			Expect(page.FindByButton("Next").Click()).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Click button 'Next' to go to disk page", func() {
			Expect(page.FindByButton("Next").Click()).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Click create virtual machine to create the vm", func() {
			Expect(page.AllByButton("Create Virtual Machine").At(1).Click()).To(Succeed())
			time.Sleep(1 * time.Second)
		})
		By("Click button 'Close' to close the wizard", func() {
			Expect(page.FindByButton("Close").Click()).To(Succeed())
			time.Sleep(1 * time.Second)
		})
	})
})
