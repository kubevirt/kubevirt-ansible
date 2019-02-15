package tests_test

import (
	"encoding/xml"
	"flag"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	launcherApi "kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
	ktests "kubevirt.io/kubevirt/tests"
)

func newPVC(pvcName, size, storageClass string) *k8sv1.PersistentVolumeClaim {
	quantity, err := resource.ParseQuantity(size)
	Expect(err).ToNot(HaveOccurred())

	name := fmt.Sprintf("%s", pvcName)

	return &k8sv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: k8sv1.PersistentVolumeClaimSpec{
			AccessModes: []k8sv1.PersistentVolumeAccessMode{k8sv1.ReadWriteOnce},
			Resources: k8sv1.ResourceRequirements{
				Requests: k8sv1.ResourceList{
					"storage": quantity,
				},
			},
			StorageClassName: &storageClass,
		},
	}
}

var _ = Describe("[rfe_id:904][crit:medium][vendor:cnv-qe@redhat.com][level:component]Verify that all disks get right 'cache' mode", func() {

	flag.Parse()

	const (
		PVName1       = "pv1"
		PVName2       = "pv2"
		PVName3       = "pv3"
		dvName        = "dv-disk"
		cirrosHttpUrl = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"

		cacheNone         = string(v1.CacheNone)
		cacheWritethrough = string(v1.CacheWriteThrough)
	)

	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	Describe("Verify disk cache mode on VM with DataVolume and glusterfs PVCs", func() {

		BeforeEach(func() {
			// Create PVCs
			_, err = virtClient.CoreV1().PersistentVolumeClaims(ktests.NamespaceTestDefault).Create(newPVC(PVName1, "1Gi", "glusterfs-storage"))
			Expect(err).ToNot(HaveOccurred())
			_, err = virtClient.CoreV1().PersistentVolumeClaims(ktests.NamespaceTestDefault).Create(newPVC(PVName2, "1Gi", "glusterfs-storage"))
			Expect(err).ToNot(HaveOccurred())
			_, err = virtClient.CoreV1().PersistentVolumeClaims(ktests.NamespaceTestDefault).Create(newPVC(PVName3, "1Gi", "glusterfs-storage"))
			Expect(err).ToNot(HaveOccurred())
		}, 60)

		It("[test_id:988]should set appropriate cache modes to VM with PVC", func() {
			dataVolume := ktests.NewRandomDataVolumeWithHttpImport(cirrosHttpUrl, ktests.NamespaceTestDefault)
			dataVolume.Name = dvName
			storageClassName := "glusterfs-storage"

			dataVolume.Spec.PVC.StorageClassName = &storageClassName
			vmi := ktests.NewRandomVMIWithDataVolume(dataVolume.Name)

			_, err := virtClient.CdiClient().CdiV1alpha1().DataVolumes(dataVolume.Namespace).Create(dataVolume)
			Expect(err).To(BeNil())

			ktests.AddPVCDisk(vmi, "pv1", "virtio", PVName1)
			vmi.Spec.Domain.Devices.Disks[1].Cache = v1.CacheNone
			ktests.AddPVCDisk(vmi, "pv2", "virtio", PVName2)
			vmi.Spec.Domain.Devices.Disks[2].Cache = v1.CacheWriteThrough
			ktests.AddPVCDisk(vmi, "pv3", "virtio", PVName3)

			ktests.RunVMIAndExpectLaunch(vmi, false, 300)

			runningVMISpec := launcherApi.DomainSpec{}
			domXml, err := ktests.GetRunningVirtualMachineInstanceDomainXML(virtClient, vmi)
			Expect(err).ToNot(HaveOccurred())
			err = xml.Unmarshal([]byte(domXml), &runningVMISpec)
			Expect(err).ToNot(HaveOccurred())

			By("checking if number of attached disks is equal to real disks number")
			disks := runningVMISpec.Devices.Disks
			Expect(len(vmi.Spec.Domain.Devices.Disks)).To(Equal(len(disks)))

			By("checking if default cache 'none' has been set to DataVolume PVC")
			Expect(disks[0].Alias.Name).To(Equal("disk0"))
			Expect(disks[0].Driver.Cache).To(Equal(cacheNone))

			By("checking if requested cache 'none' has been set to PVC")
			Expect(disks[1].Alias.Name).To(Equal("pv1"))
			Expect(disks[1].Driver.Cache).To(Equal(cacheNone))

			By("checking if requested cache 'writethrough' has been set to PVC")
			Expect(disks[2].Alias.Name).To(Equal("pv2"))
			Expect(disks[2].Driver.Cache).To(Equal(cacheWritethrough))

			By("checking if default cache 'none' has been set to PVC disk")
			Expect(disks[3].Alias.Name).To(Equal("pv3"))
			Expect(disks[3].Driver.Cache).To(Equal(cacheNone))
		})
	})

})
