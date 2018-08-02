package tests

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/tests"
)

var _ = Describe("Preset liftcycle", func() {
	flag.Parse()

	_, err := kubecli.GetKubevirtClient()
	tests.PanicOnError(err)

	var vmPreset *v1.VirtualMachineInstancePreset
	var presetJsonFile string

	flavorKey := fmt.Sprintf("%s/flavor", v1.GroupName)
	memoryFlavor := "memory-test"
	memory, _ := resource.ParseQuantity("64M")

	BeforeEach(func() {
		tests.SkipIfNoOc()
		tests.BeforeTestCleanup()

		selector := k8smetav1.LabelSelector{MatchLabels: map[string]string{flavorKey: memoryFlavor}}
		vmPreset = &v1.VirtualMachineInstancePreset{
			TypeMeta:   k8smetav1.TypeMeta{APIVersion: "kubevirt.io/v1alpha2", Kind: "VirtualMachineInstancePreset"},
			ObjectMeta: k8smetav1.ObjectMeta{Name: "test-preset"},
			Spec: v1.VirtualMachineInstancePresetSpec{
				Selector: selector,
				Domain: &v1.DomainPresetSpec{
					Resources: v1.ResourceRequirements{Requests: k8sv1.ResourceList{
						"memory": memory}},
				},
			},
		}
	})

	Describe("VMIPreset", func() {

		assertGeneratedPresetJson := func() func() {
			return func() {
				By("Generating Preset JSON file")
				_, err := generatePresetJson(vmPreset)
				Expect(err).ToNot(HaveOccurred())
				Expect(presetJsonFile).To(BeAnExistingFile())
			}
		}

		assertCreatedPreset := func() func() {
			return func() {
				By("Creating preset via oc command")
				out, err := runOcCreateCommand(presetJsonFile)
				ExpectWithOffset(1, err).ToNot(HaveOccurred())
				message := fmt.Sprintf("virtualmachineinstancepreset.kubevirt.io \"%s\" created\n", vmPreset.Name)
				Expect(out).To(Equal(message))
			}
		}

		assertDeletedPreset := func() func() {
			return func() {
				By("Deleting preset via oc command")
				out, err := runOcDeletePresetCommand(vmPreset.Name)
				ExpectWithOffset(1, err).ToNot(HaveOccurred())
				message := fmt.Sprintf("virtualmachineinstancepreset.kubevirt.io \"%s\" deleted\n", vmPreset.Name)
				ExpectWithOffset(1, out).To(Equal(message))
			}
		}

		assertGetPreset := func() func() {
			return func() {
				By("Checking preset exists via oc command")
				EventuallyWithOffset(1, func() bool {
					out, err := runOcGetCommand("virtualmachineinstancepreset.kubevirt.io")
					ExpectWithOffset(1, err).ToNot(HaveOccurred())
					return strings.Contains(out, vmPreset.Name)
				}, time.Duration(45)*time.Second).Should(BeTrue(), "Timed out waiting for preset to appear")
			}
		}

		assertRemovedFile := func(file string) func() {
			return func() {
				if _, err := os.Stat(file); !os.IsNotExist(err) {
					err := os.Remove(file)
					ExpectWithOffset(1, err).ToNot(HaveOccurred())
				}
				ExpectWithOffset(1, file).NotTo(BeAnExistingFile())
			}
		}

		testGivenPreset := func() {

			Context("Preset testing", func() {

				It("should succeed to generate a Preset", assertGeneratedPresetJson())

				It("should succeed to create preset using oc command", assertCreatedPreset())

				It("should succeed to get preset using oc command", assertGetPreset())

				It("should succeed to delete the preset using oc command", assertDeletedPreset())
			})
		}

		BeforeEach(func() {
			presetJsonFile = fmt.Sprintf("%s.json", vmPreset.Name)
			Expect(presetJsonFile).NotTo(BeAnExistingFile())
		})

		JustBeforeEach(func() {
			var err error
			presetJsonFile, err = generatePresetJson(vmPreset)
			Expect(err).ToNot(HaveOccurred())
			Expect(presetJsonFile).To(BeAnExistingFile())
		})

		AfterEach(func() {
			assertRemovedFile(presetJsonFile)()
		})

		Context("preset life-cycle testing", func() {

			testGivenPreset()
		})
	})
})

func generatePresetJson(preset *v1.VirtualMachineInstancePreset) (string, error) {
	data, err := json.Marshal(preset)
	if err != nil {
		return "", fmt.Errorf("failed to generate json for preset %s", preset.Name)
	}

	jsonFile := fmt.Sprintf("%s.json", preset.Name)
	err = ioutil.WriteFile(jsonFile, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write json file %s", jsonFile)
	}
	return jsonFile, nil
}

func runOcCreateCommand(presetJsonFile string) (string, error) {
	out, err := tests.RunOcCommand("create", "-f", presetJsonFile)
	return out, err
}

func runOcDeletePresetCommand(presetName string) (string, error) {
	out, err := tests.RunOcCommand("delete", "VirtualMachineInstancePreset", presetName)
	return out, err
}

func runOcGetCommand(resourceType string) (string, error) {
	out, err := tests.RunOcCommand("get", resourceType)
	return out, err
}
