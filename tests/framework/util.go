package framework

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"kubevirt.io/kubevirt/pkg/api/v1"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
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

func GetResourceSpecificParameters(resourceType, resourceName, query string) string {
	return execute(Result{cmd: "oc", verb: "get", resourceType: resourceType, resourceName: resourceName, query: query})
}

func WaitUntilResourceDeleted(resourceType, resourceName string) {
	Eventually(func() bool {
		res, _ := GetObjects(NamespaceTestDefault, resourceType)
		return !strings.Contains(strings.Join(res, ""), resourceName)
	}, LongTimeout).Should(BeTrue(), fmt.Sprintf("Timed out waiting for %s ", resourceType))
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

func AddСPU(vmi *v1.VirtualMachineInstance, cores uint32, cpuModel string) *v1.VirtualMachineInstance {
	vmi.Spec.Domain.CPU = &v1.CPU{
		Cores: cores,
		Model: cpuModel,
	}

	// workaround of bug (fedora 29 does not boot well with Nehalem CPU model)
	if cpuModel == "Nehalem" {
		vmi.Spec.Domain.Devices.Rng = &v1.Rng{}
	}

	return vmi
}

func CreatePVC(PVCName, size, storageClass string, accessMode k8sv1.PersistentVolumeAccessMode) {
	virtCli, err := kubecli.GetKubevirtClient()
	Expect(err).ToNot(HaveOccurred())

	_, err = virtCli.CoreV1().PersistentVolumeClaims(NamespaceTestDefault).Create(newPVC(PVCName, size, storageClass, accessMode))
	if !errors.IsAlreadyExists(err) {
		Expect(err).ToNot(HaveOccurred())
	}
}

func newPVC(PVCName, size, storageClass string, accessMode k8sv1.PersistentVolumeAccessMode) *k8sv1.PersistentVolumeClaim {
	quantity, err := resource.ParseQuantity(size)
	Expect(err).ToNot(HaveOccurred())

	return &k8sv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: PVCName},
		Spec: k8sv1.PersistentVolumeClaimSpec{
			AccessModes: []k8sv1.PersistentVolumeAccessMode{accessMode},
			Resources: k8sv1.ResourceRequirements{
				Requests: k8sv1.ResourceList{
					"storage": quantity,
				},
			},
			StorageClassName: &storageClass,
		},
	}
}

func CreateNFSPvAndPvc(PVName, size, path, server string, accessMode k8sv1.PersistentVolumeAccessMode) {
	virtCli, err := kubecli.GetKubevirtClient()
	Expect(err).ToNot(HaveOccurred())

	_, err = virtCli.CoreV1().PersistentVolumes().Create(newNFSpv(PVName, size, path, server, accessMode))
	if !errors.IsAlreadyExists(err) {
		Expect(err).ToNot(HaveOccurred())
	}

	_, err = virtCli.CoreV1().PersistentVolumeClaims(NamespaceTestDefault).Create(newNFSpvc(PVName, size, accessMode))
	if !errors.IsAlreadyExists(err) {
		Expect(err).ToNot(HaveOccurred())
	}
}

func newNFSpv(PVName, size, path, server string, accessMode k8sv1.PersistentVolumeAccessMode) *k8sv1.PersistentVolume {
	quantity, err := resource.ParseQuantity(size)
	Expect(err).ToNot(HaveOccurred())

	storageClass := StorageClassLocal

	return &k8sv1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: PVName,
		},
		Spec: k8sv1.PersistentVolumeSpec{
			AccessModes: []k8sv1.PersistentVolumeAccessMode{accessMode},
			Capacity: k8sv1.ResourceList{
				"storage": quantity,
			},
			ClaimRef: &k8sv1.ObjectReference{
				Name:      PVName,
				Namespace: NamespaceTestDefault,
			},
			StorageClassName:              storageClass,
			PersistentVolumeReclaimPolicy: k8sv1.PersistentVolumeReclaimRecycle,
			PersistentVolumeSource: k8sv1.PersistentVolumeSource{
				NFS: &k8sv1.NFSVolumeSource{
					Server: server,
					Path:   path,
				},
			},
		},
	}
}

func newNFSpvc(name string, size string, accessMode k8sv1.PersistentVolumeAccessMode) *k8sv1.PersistentVolumeClaim {
	quantity, err := resource.ParseQuantity(size)
	Expect(err).ToNot(HaveOccurred())

	storageClass := StorageClassLocal

	return &k8sv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: k8sv1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			AccessModes:      []k8sv1.PersistentVolumeAccessMode{accessMode},
			Resources: k8sv1.ResourceRequirements{
				Requests: k8sv1.ResourceList{
					"storage": quantity,
				},
			},
		},
	}
}
