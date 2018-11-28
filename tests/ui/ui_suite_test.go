package ui_test

import (
	"flag"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)

func TestUi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ui Suite")
}

const uinamespace = "kubevirt-web-ui"
const url = "https://kubevirt-web-ui.cloudapps.example.com"
const username = "test_admin"
const password = "123456"
const timeout = 60 * time.Second
const pollInterval = 2 * time.Second

var agoutiDriver *agouti.WebDriver

var webDriver string

var headless bool

func init() {
	flag.BoolVar(&headless, "headless", true, "run the test in headless mode")
	flag.StringVar(&webDriver, "webDriver", "", "webDriver used for test, eg: chrome, firefox")
}

var _ = BeforeSuite(func() {
	flag.Parse()

	capabilities := agouti.NewCapabilities().With("acceptInsecureCerts")
	switch webDriver {
	case "chromedriver":
		if headless {
			agoutiDriver = agouti.ChromeDriver(
				agouti.ChromeOptions("args", []string{
					"--headless",
					"--allow-insecure-localhost",
					"--disable-gpu",
				}),
				agouti.Debug,
				agouti.Desired(capabilities),
			)
		} else {
			agoutiDriver = agouti.ChromeDriver(
				agouti.Debug,
				agouti.Desired(capabilities),
			)
		}
	case "geckodriver":
		if headless {
			agoutiDriver = agouti.GeckoDriver(agouti.Debug)
		} else {
			agoutiDriver = agouti.GeckoDriver(agouti.Debug)
		}
	case "selenium":
		command := []string{"java", "-jar", "/usr/local/bin/selenium-server-standalone-3.14.0.jar", "-port", "{{.Port}}"}
		agoutiDriver = agouti.NewWebDriver("http://{{.Address}}/wd/hub", command, agouti.Debug)
		agoutiDriver = agouti.Selenium()
	default:
		panic("webDriver must not be empty and it should be chrome or firefox")
	}

	Expect(agoutiDriver.Start()).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(agoutiDriver.Stop()).To(Succeed())
})
