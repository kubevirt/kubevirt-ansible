/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package tests_test

import (
	"flag"
	"os/exec"
	"time"
	"encoding/xml"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/tests"
)

var _ = Describe("IOThreads", func() {
	flag.Parse()

	virtClient, err := kubecli.GetKubevirtClient()
	tests.PanicOnError(err)

	autoPolicy := v1.IOThreadsPolicyAuto
	_true := true
	IOThreadSpec := v1.VirtualMachineInstanceSpec{
	    Domain: v1.DomainSpec{
		Resources: v1.ResourceRequirements{
		    Requests: k8sv1.ResourceList{
			k8sv1.ResourceMemory: resource.MustParse("1024M"),
			k8sv1.ResourceCPU: resource.MustParse("2"),
		    },
		},
		IOThreadsPolicy: &autoPolicy,

		Devices: v1.Devices{
		    Disks: []v1.Disk{
			{
			    Name:       "ded1",
			    VolumeName: "ded1volume",
			    DiskDevice: v1.DiskDevice{
				Disk: &v1.DiskTarget{
				    Bus: "virtio",
				},
			    },
			    DedicatedIOThread: &_true,
			},
			{
			    Name:       "ded2",
			    VolumeName: "ded2volume",
			    DiskDevice: v1.DiskDevice{
				Disk: &v1.DiskTarget{
				    Bus: "virtio",
				},
			    },
			    DedicatedIOThread: &_true,
			},
			{
			    Name:       "shr1",
			    VolumeName: "shr1volume",
			    DiskDevice: v1.DiskDevice{
				Disk: &v1.DiskTarget{
				    Bus: "virtio",
				},
			    },
			},
			{
			    Name:       "shr2",
			    VolumeName: "shr2volume",
			    DiskDevice: v1.DiskDevice{
				Disk: &v1.DiskTarget{
				    Bus: "virtio",
				},
			    },
			},
			{
			    Name:       "shr3",
			    VolumeName: "shr3volume",
			    DiskDevice: v1.DiskDevice{
				Disk: &v1.DiskTarget{
				    Bus: "virtio",
				},
			    },
			},
			{
			    Name:       "shr4",
			    VolumeName: "shr4volume",
			    DiskDevice: v1.DiskDevice{
				Disk: &v1.DiskTarget{
				    Bus: "virtio",
				},
			    },
			},
		    },
		},
	    },
	    Volumes: []v1.Volume{
		{
		    Name: "ded1volume",
		    VolumeSource: v1.VolumeSource{
			EmptyDisk: &v1.EmptyDiskSource{
			    Capacity: resource.MustParse("1G"),
			},
		    },
		},
		{
		    Name: "ded2volume",
		    VolumeSource: v1.VolumeSource{
			EmptyDisk: &v1.EmptyDiskSource{
			    Capacity: resource.MustParse("1G"),
			},
		    },
		},
		{
		    Name: "shr1volume",
		    VolumeSource: v1.VolumeSource{
			EmptyDisk: &v1.EmptyDiskSource{
			    Capacity: resource.MustParse("1G"),
			},
		    },
		},
		{
		    Name: "shr2volume",
		    VolumeSource: v1.VolumeSource{
			EmptyDisk: &v1.EmptyDiskSource{
			    Capacity: resource.MustParse("1G"),
			},
		    },
		},
		{
		    Name: "shr3volume",
		    VolumeSource: v1.VolumeSource{
			EmptyDisk: &v1.EmptyDiskSource{
			    Capacity: resource.MustParse("1G"),
			},
		    },
		},
		{
		    Name: "shr4volume",
		    VolumeSource: v1.VolumeSource{
			EmptyDisk: &v1.EmptyDiskSource{
			    Capacity: resource.MustParse("1G"),
			},
		    },
		},
	    },
	}

	var IOThreadVMI *v1.VirtualMachineInstance

	BeforeEach(func() {
		tests.BeforeTestCleanup()
		IOThreadVMI = tests.NewRandomVMI()
		IOThreadVMI.Spec = IOThreadSpec
	})

	Context("IOThreads Policies - auto x2", func() {

		It("Virtual Disk Settings - IOThreads", func() {

			// How many symbols in the names of pod and VMI shoul match
			symbols_to_compare := 30

			By("Creating VMI with desired spec")
			IOThreadVMI, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(IOThreadVMI)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(IOThreadVMI.Name) > symbols_to_compare).To(BeTrue(), "VMI Name should contain at least N symbols")

			By("Checking that corresponding pod exists")
			listOptions := metav1.ListOptions{}
			podList, err := virtClient.CoreV1().Pods(tests.NamespaceTestDefault).List(listOptions)
			Expect(err).ToNot(HaveOccurred())
			Expect(podList.Items).To(HaveLen(1), "We should only have 1 pod")
			// I don't know how exactly VM and names are formed, and as a result I can't properly compare 
			// VM names with pod names, since they both have a number or random symbols at the end,
			// and these symbols are different for pod name and vm name. So, I only can compare their names before
			// these random symbols.
			Expect(podList.Items[0].Name).To(HavePrefix("virt-launcher-" + IOThreadVMI.Name[:symbols_to_compare]), "Pod name should have a name similiar to VM name")

			By("Checking that VMI with this name does exist")
			getOptions := metav1.GetOptions{}
			resultVMI, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Get(IOThreadVMI.Name, &getOptions)
			Expect(err).ToNot(HaveOccurred())

			Expect(*resultVMI.Spec.Domain.IOThreadsPolicy).To(Equal(autoPolicy), "Spec should have auto policy")

			By("Checking that spec has correct dedicated threads")
			ded_n := 0
			for _, disk := range resultVMI.Spec.Domain.Devices.Disks {
				if strings.HasPrefix(disk.Name, "ded") {
					Expect(*disk.DedicatedIOThread).To(BeTrue(), "Should have dedicated IO thread")
					ded_n += 1
				}
			}
			Expect(ded_n).To(Equal(2), "There should be 2 disks with dedicated threads")

			By("Checking that exported XML has correct dedicated threads")
			duration := time.Duration(60)*time.Second
			time.Sleep(duration)
			//tests.WaitUntilVMIReadyWithNamespace(tests.NamespaceTestDefault, resultVMI, tests.LoggedInCirrosExpecter)
			command := "/usr/local/bin/oc project kubevirt-test-default && "
			command += "/usr/local/bin/kubectl"
			command += " exec " + podList.Items[0].Name
			command += " --container compute cat"
			command += " /var/run/libvirt/qemu/kubevirt-test-default_" + IOThreadVMI.Name + ".xml"
			output, err := exec.Command("/bin/bash", "-c", command).Output()
			type DomStatus struct {
				Domain api.DomainSpec	`xml:"domain"`
			}
			domStat := &DomStatus{}
			err = xml.Unmarshal(output, domStat)
			Expect(err).ToNot(HaveOccurred())

			var threads [5]uint
			for _, disk := range domStat.Domain.Devices.Disks {
				thread_n := uint(*disk.Driver.IOThread)
				// Just checking to avod addressing outside of array boundaries
				Expect(uint(*disk.Driver.IOThread) < 5).To(BeTrue())
				if strings.HasPrefix(disk.Alias.Name, "ded") {
					// Dedicated thread numbers should be unique, i.e. they shouldn't be encountered before
					// If we already encountered thread with this number - this value will be > 0
					Expect(threads[thread_n] == 0).To(BeTrue(), "Dedicated thread number should be unique")
				}
				threads[uint(*disk.Driver.IOThread)] += 1
			}
			Expect(threads[1] == 2 && threads[2] == 2).To(BeTrue(), "Threads 1 and 2 should be shared")
			Expect(threads[3] == 1 && threads[4] == 1).To(BeTrue(), "Threads 3 and 4 should be dedicated")
			Expect(domStat.Domain.IOThreads.IOThreads == 4).To(BeTrue(), "There should be 4 iothreads")
		})
	})

	Context("IOThreads Policies - auto x3", func() {

		It("Virtual Disk Settings - IOThreads", func() {

			IOThreadVMI.Spec.Domain.Resources.Requests = k8sv1.ResourceList{
				k8sv1.ResourceMemory: resource.MustParse("1024M"),
				k8sv1.ResourceCPU: resource.MustParse("3"),
			}
			// How many symbols in the names of pod and VMI shoul match
			symbols_to_compare := 30

			By("Creating VMI with desired spec")
			IOThreadVMI, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(IOThreadVMI)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(IOThreadVMI.Name) > symbols_to_compare).To(BeTrue(), "VMI Name should contain at least N symbols")

			By("Checking that corresponding pod exists")
			listOptions := metav1.ListOptions{}
			podList, err := virtClient.CoreV1().Pods(tests.NamespaceTestDefault).List(listOptions)
			Expect(err).ToNot(HaveOccurred())
			Expect(podList.Items).To(HaveLen(1), "We should only have 1 pod")
			// I don't know how exactly VM and names are formed, and as a result I can't properly compare 
			// VM names with pod names, since they both have a number or random symbols at the end,
			// and these symbols are different for pod name and vm name. So, I only can compare their names before
			// these random symbols.
			Expect(podList.Items[0].Name).To(HavePrefix("virt-launcher-" + IOThreadVMI.Name[:symbols_to_compare]), "Pod name should have a name similiar to VM name")

			By("Checking that VMI with this name does exist")
			getOptions := metav1.GetOptions{}
			resultVMI, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Get(IOThreadVMI.Name, &getOptions)
			Expect(err).ToNot(HaveOccurred())

			Expect(*resultVMI.Spec.Domain.IOThreadsPolicy).To(Equal(autoPolicy), "Spec should have auto policy")

			By("Checking that spec has correct dedicated threads")
			ded_n := 0
			for _, disk := range resultVMI.Spec.Domain.Devices.Disks {
				if strings.HasPrefix(disk.Name, "ded") {
					Expect(*disk.DedicatedIOThread).To(BeTrue(), "Should have dedicated IO thread")
					ded_n += 1
				}
			}
			Expect(ded_n).To(Equal(2), "There should be 2 disks with dedicated threads")

			By("Checking that exported XML has correct dedicated threads")
			duration := time.Duration(60)*time.Second
			time.Sleep(duration)
			//tests.WaitUntilVMIReadyWithNamespace(tests.NamespaceTestDefault, resultVMI, tests.LoggedInCirrosExpecter)
			command := "/usr/local/bin/oc project kubevirt-test-default && "
			command += "/usr/local/bin/kubectl"
			command += " exec " + podList.Items[0].Name
			command += " --container compute cat"
			command += " /var/run/libvirt/qemu/kubevirt-test-default_" + IOThreadVMI.Name + ".xml"
			output, err := exec.Command("/bin/bash", "-c", command).Output()
			type DomStatus struct {
				Domain api.DomainSpec	`xml:"domain"`
			}
			domStat := &DomStatus{}
			err = xml.Unmarshal(output, domStat)
			Expect(err).ToNot(HaveOccurred())

			var threads [7]uint
			for _, disk := range domStat.Domain.Devices.Disks {
				thread_n := uint(*disk.Driver.IOThread)
				// Just checking to avod addressing outside of array boundaries
				Expect(uint(*disk.Driver.IOThread) < 7).To(BeTrue())
				if strings.HasPrefix(disk.Alias.Name, "ded") {
					// Dedicated thread numbers should be unique, i.e. they shouldn't be encountered before
					// If we already encountered thread with this number - this value will be > 0
					Expect(threads[thread_n] == 0).To(BeTrue(), "Dedicated thread number should be unique")
				}
				threads[uint(*disk.Driver.IOThread)] += 1
			}
			Expect(threads[1] == 1 && threads[2] == 1 && threads[3] == 1 && threads[4] == 1).To(BeTrue(), "Threads 1-4 should be shared but still unique since we have enough threads")
			Expect(threads[5] == 1 && threads[6] == 1).To(BeTrue(), "Threads 5 and 6 should be dedicated")
			Expect(domStat.Domain.IOThreads.IOThreads == 6).To(BeTrue(), "There should be 6 iothreads")
		})
	})
})

