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

var _ = Describe("[rfe_id:384][crit:medium][vendor:cnv-qe@redhat.com][level:component]Config", func() {

	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	Context("With a Secret and a ConfigMap defined", func() {

		Context("With a single volume", func() {
			var (
				configMapName string
				configMapPath string
				secretName    string
				secretPath    string
			)

			BeforeEach(func() {
				configMapName = "configmap-" + uuid.NewRandom().String()
				configMapPath = config.GetConfigMapSourcePath(configMapName + "-disk")
				secretName = "secret-" + uuid.NewRandom().String()
				secretPath = config.GetSecretSourcePath(secretName + "-disk")

				config_data := map[string]string{
					"config1": "value1",
					"config2": "value2",
					"config3": "value3",
				}

				secret_data := map[string]string{
					"user":     "admin",
					"password": "redhat",
				}

				ktests.CreateConfigMap(configMapName, config_data)

				ktests.CreateSecret(secretName, secret_data)
			})

			AfterEach(func() {
				ktests.DeleteConfigMap(configMapName)
				ktests.DeleteSecret(secretName)
			})

			It("[test_id:786]Should be that cfgMap and secret fs layout same for the pod and vmi", func() {
				expectedOutput_cfgMap := "value1value2value3"
				expectedOutput_Secret := "adminredhat"

				By("Running VMI")

				vmi := ktests.NewRandomVMIWithEphemeralDiskAndUserdataHighMemory(
					ktests.ContainerDiskFor(
						ktests.ContainerDiskFedora), "#!/bin/bash\necho \"fedora\" | passwd fedora --stdin\n")
				ktests.AddConfigMapDisk(vmi, configMapName)
				ktests.AddSecretDisk(vmi, secretName)
				ktests.RunVMIAndExpectLaunch(vmi, false, 180)

				By("Checking if ConfigMap has been attached to the pod")
				vmiPod := ktests.GetRunningPodByVirtualMachineInstance(vmi, tests.NamespaceTestDefault)
				podOutput_cfgMap, err := ktests.ExecuteCommandOnPod(
					virtClient,
					vmiPod,
					vmiPod.Spec.Containers[1].Name,
					[]string{"cat",
						configMapPath + "/config1",
						configMapPath + "/config2",
						configMapPath + "/config3",
					},
				)
				Expect(err).To(BeNil())
				Expect(podOutput_cfgMap).To(Equal(expectedOutput_cfgMap))

				By("Checking mounted ConfigMap image")
				expecter, err := tests.LoggedInFedoraExpecter(vmi.Name, tests.NamespaceTestDefault, 360)
				Expect(err).ToNot(HaveOccurred())
				defer expecter.Close()

				_, err = expecter.ExpectBatch([]expect.Batcher{
					// mount ConfigMap image
					&expect.BSnd{S: "sudo su -\n"},
					&expect.BExp{R: "#"},
					&expect.BSnd{S: "mount /dev/sda /mnt\n"},
					&expect.BSnd{S: "echo $?\n"},
					&expect.BExp{R: "0"},
					&expect.BSnd{S: "cat /mnt/config1 /mnt/config2 /mnt/config3\n"},
					&expect.BExp{R: expectedOutput_cfgMap},
				}, 200*time.Second)
				Expect(err).ToNot(HaveOccurred())

				By("Checking if Secret has also been attached to the same pod")
				podOutput_Secret, err := ktests.ExecuteCommandOnPod(
					virtClient,
					vmiPod,
					vmiPod.Spec.Containers[1].Name,
					[]string{"cat",
						secretPath + "/user",
						secretPath + "/password",
					},
				)
				Expect(err).To(BeNil())
				Expect(podOutput_Secret).To(Equal(expectedOutput_Secret))

				By("Checking mounted secret image")

				_, err = expecter.ExpectBatch([]expect.Batcher{
					// mount Secret image
					&expect.BSnd{S: "mount /dev/sdb /mnt\n"},
					&expect.BSnd{S: "echo $?\n"},
					&expect.BExp{R: "0"},
					&expect.BSnd{S: "cat /mnt/user /mnt/password\n"},
					&expect.BExp{R: expectedOutput_Secret},
				}, 200*time.Second)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("With SSH Keys as a Secret defined", func() {

		Context("With a single volume", func() {
			var (
				secretName string
				secretPath string
			)

			var bitSize int = 2048
			privateKey, _ := tests.GeneratePrivateKey(bitSize)
			publicKeyBytes, _ := tests.GeneratePublicKey(&privateKey.PublicKey)
			privateKeyBytes := tests.EncodePrivateKeyToPEM(privateKey)

			BeforeEach(func() {
				secretName = "secret-" + uuid.NewRandom().String()
				secretPath = config.GetSecretSourcePath(secretName + "-disk")

				data := map[string]string{
					"ssh-privatekey": string(privateKeyBytes),
					"ssh-publickey":  string(publicKeyBytes),
				}
				ktests.CreateSecret(secretName, data)
			})

			AfterEach(func() {
				ktests.DeleteSecret(secretName)
			})

			It("[test_id:778]Should be the fs layout the same for a pod and vmi", func() {
				expectedPrivateKey := string(privateKeyBytes)
				expectedPublicKey := string(publicKeyBytes)

				By("Running VMI")
				vmi := ktests.NewRandomVMIWithEphemeralDiskAndUserdataHighMemory(
					ktests.ContainerDiskFor(
						ktests.ContainerDiskFedora), "#!/bin/bash\necho \"fedora\" | passwd fedora --stdin\n")
				ktests.AddSecretDisk(vmi, secretName)
				ktests.RunVMIAndExpectLaunch(vmi, false, 180)

				By("Checking if Secret has been attached to the pod")
				vmiPod := ktests.GetRunningPodByVirtualMachineInstance(vmi, tests.NamespaceTestDefault)
				podOutput1, err := ktests.ExecuteCommandOnPod(
					virtClient,
					vmiPod,
					vmiPod.Spec.Containers[1].Name,
					[]string{"cat",
						secretPath + "/ssh-privatekey",
					},
				)
				Expect(err).To(BeNil())
				Expect(podOutput1).To(Equal(expectedPrivateKey))

				podOutput2, err := ktests.ExecuteCommandOnPod(
					virtClient,
					vmiPod,
					vmiPod.Spec.Containers[1].Name,
					[]string{"cat",
						secretPath + "/ssh-publickey",
					},
				)
				Expect(err).To(BeNil())
				Expect(podOutput2).To(Equal(expectedPublicKey))

				By("Checking mounted secrets sshkeys image")
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
					&expect.BSnd{S: "grep \"PRIVATE KEY\" /mnt/ssh-privatekey\n"},
					&expect.BSnd{S: "echo $?\n"},
					&expect.BExp{R: "0"},
					&expect.BSnd{S: "grep ssh-rsa /mnt/ssh-publickey\n"},
					&expect.BSnd{S: "echo $?\n"},
					&expect.BExp{R: "0"},
				}, 200*time.Second)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
