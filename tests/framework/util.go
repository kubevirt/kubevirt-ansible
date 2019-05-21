package framework

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	randk8s "k8s.io/apimachinery/pkg/util/rand"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
	"math"
	"net/http"
	"os/exec"
	"strings"
)

func ProcessTemplateWithParameters(srcFilePath, dstFilePath string, params ...string) string {
	By(fmt.Sprintf("Overriding the template from %s to %s", srcFilePath, dstFilePath))
	out := execute(Result{cmd: "oc", verb: "process", filePath: srcFilePath, params: params})
	filePath, err := writeJson(dstFilePath, out)
	Expect(err).ToNot(HaveOccurred())
	return filePath
}

func CreateResourceWithFilePathTestNamespace(filePath string) {
	By("Creating resource from the json file with the oc-create command")
	execute(Result{cmd: "oc", verb: "create", filePath: filePath})
}

func DeleteResourceWithLabelTestNamespace(resourceType, resourceLabel string) {
	By(fmt.Sprintf("Deleting %s:%s from the json file with the oc-delete command", resourceType, resourceLabel))
	execute(Result{cmd: "oc", verb: "delete", resourceType: resourceType, resourceLabel: resourceLabel})
}

func DeleteResourceByName(resourceType, nameSpace, resourceName string) {
	By(fmt.Sprintf("Deleting %s:%s  from %s with oc-delete command", resourceType, resourceName, nameSpace))
	execute(Result{cmd: "oc", verb: "delete", resourceType: resourceType, nameSpace: nameSpace, resourceName: resourceName})
}

func CreateResourceWithFilePath(filePath string) {
	By("Creating resource from the json file with the oc-create command")
	execute(Result{cmd: "oc", verb: "create", filePath: filePath})
}

func WaitUntilResourceReadyByNameTestNamespace(resourceType, resourceName, query, expectOut string) {
	By(fmt.Sprintf("Wait until %s with name %s ready", resourceType, resourceName))
	execute(Result{cmd: "oc", verb: "get", resourceType: resourceType, resourceName: resourceName, query: query, expectOut: expectOut})
}

func WaitUntilResourceReadyByName(resourceType, resourceName, nameSpace, query, expectOut string) {
	By(fmt.Sprintf("Wait until %s with name %s ready", resourceType, resourceName))
	execute(Result{cmd: "oc", verb: "get", resourceType: resourceType, resourceName: resourceName, nameSpace: nameSpace, query: query, expectOut: expectOut})
}

func WaitUntilResourceReadyByLabelTestNamespace(resourceType, label, query, expectOut string) {
	By(fmt.Sprintf("Wait until resource %s with label=%s ready", resourceType, label))
	execute(Result{cmd: "oc", verb: "get", resourceType: resourceType, resourceLabel: label, query: query, expectOut: expectOut})
}

func writeJson(jsonFile string, json string) (string, error) {
	err := ioutil.WriteFile(jsonFile, []byte(json), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write the json file %s", jsonFile)
	}
	return jsonFile, nil
}

func RunOcDescribeCommand(resourceType, resourceName string) string {
	fmt.Printf("Getting 'oc describe' with: %s ", resourceName)
	return execute(Result{cmd: "oc", verb: "describe", resourceType: resourceType, resourceName: resourceName})
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func GeneratePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// Will returns in the format "ssh-rsa ..."
func GeneratePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	return publicKeyBytes, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privateBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privateBlock)

	return privatePEM
}

func CreateServiceAccount(saName string) {
	virtCli, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	sa := k8sv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: NamespaceTestDefault,
		},
	}

	_, err = virtCli.CoreV1().ServiceAccounts(NamespaceTestDefault).Create(&sa)
	Expect(err).ToNot(HaveOccurred())
}

func DeleteServiceAccount(saName string) {
	virtCli, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	err = virtCli.CoreV1().ServiceAccounts(NamespaceTestDefault).Delete(saName, nil)
	Expect(err).ToNot(HaveOccurred())
}

func RemoveDataVolume(dvName string, namespace string) {
	virtCli, err := kubecli.GetKubevirtClient()
	Expect(err).ToNot(HaveOccurred())
	err = virtCli.CdiClient().CdiV1alpha1().DataVolumes(namespace).Delete(dvName, nil)
	Expect(err).ToNot(HaveOccurred())
}


func GetAvailableResources(virtClient kubecli.KubevirtClient, cpuNeeded int64, memNeeded int64) (int, int) {
	nodeList := ktests.GetAllSchedulableNodes(virtClient)
	cpu_limit_total, mem_limit_total := 0, 0
	for _, node := range nodeList.Items {
		cpu := node.Status.Allocatable["cpu"]
		mem := node.Status.Allocatable["memory"]
		available_cpu, CpuOK := (&cpu).AsInt64()
		available_mem, MemOK := (&mem).AsInt64()
		if CpuOK && MemOK {
			cpu_limit := int(available_cpu / cpuNeeded)
			mem_limit := int(available_mem / memNeeded)
			cpu_limit_total += cpu_limit
			mem_limit_total += mem_limit
		}
	}

	return cpu_limit_total, mem_limit_total
}

// Checking if the cluster can run at least one VM
func isEnoughResources(virtClient kubecli.KubevirtClient, cpuNeeded int, memNeeded int64) (bool, int, int) {
	availableVMs, cpu_limit, mem_limit := GetAvailableResources(virtClient, int64(cpuNeeded), memNeeded)
	if availableVMs == 0 {
		return false, cpu_limit, mem_limit

	} else {
		return true, cpu_limit, mem_limit

	}
}

func GetLatestGitHubReleaseURL(user_name string, repo_name string) string {
	github_api_address := "https://api.github.com/repos/" + user_name + "/" + repo_name + "/releases/latest"
	url_byte, err := exec.Command("/bin/bash", "-c", "curl -s "+github_api_address+" | grep browser_download_url | cut -d '\"' -f 4").Output()
	ktests.PanicOnError(err)
	return strings.TrimSuffix(string(url_byte), "\n")
}
func DownloadFile(file_url string) []byte {
	response, err := http.Get(file_url)
	ktests.PanicOnError(err)
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	ktests.PanicOnError(err)
	return data
}

func CreatingWinRmiPod(winrmCliName string, virtClient kubecli.KubevirtClient) *k8sv1.Pod {
	var winrmcliPod *k8sv1.Pod
	winrmcliPod = &k8sv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: winrmCliName + randk8s.String(5)},
		Spec: k8sv1.PodSpec{
			Containers: []k8sv1.Container{
				{
					Name:    winrmCliName,
					Image:   fmt.Sprintf("%s/%s:%s", ktests.KubeVirtRepoPrefix, winrmCliName, ktests.KubeVirtVersionTag),
					Command: []string{"sleep"},
					Args:    []string{"3600"},
				},
			},
		},
	}
	winrmcliPod, err := virtClient.CoreV1().Pods(ktests.NamespaceTestDefault).Create(winrmcliPod)
	Expect(err).ToNot(HaveOccurred())
	return winrmcliPod
}

func RunPsCommandInWindowsVM(winrmCliCmd, ip, username, password, command string, virtClient kubecli.KubevirtClient, winrmcliPod *k8sv1.Pod) (string, bool) {
	cli := []string{winrmCliCmd, "-hostname", ip, "-username", username, "-password", password}
	resultCommand := append(cli, command)
	fmt.Println("command=", resultCommand)

	By(fmt.Sprintf("Running \"%s\" command via winrm-cli", resultCommand))

	output, err := ktests.ExecuteCommandOnPod(virtClient, winrmcliPod, winrmcliPod.Spec.Containers[0].Name, resultCommand)
	if err != nil {
		return output, false
	}
	return output, true

}
