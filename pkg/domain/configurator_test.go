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
 * Copyright 2025 Red Hat, Inc.
 *
 */

package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	vmschema "kubevirt.io/api/core/v1"
	libvirtxml "libvirt.org/go/libvirtxml"

	"kubevirt.io/vhostuser-network-binding-plugin/pkg/domain"
	"kubevirt.io/vhostuser-network-binding-plugin/pkg/utils"
)

// pciAddr is a helper to build a DomainAddressPCI from four uint values.
func pciAddr(dom, bus, slot, fn uint) *libvirtxml.DomainAddressPCI {
	return &libvirtxml.DomainAddressPCI{
		Domain:   &dom,
		Bus:      &bus,
		Slot:     &slot,
		Function: &fn,
	}
}

var _ = Describe("vhostuser network configurator", func() {
	Context("generate domain spec interface", func() {
		DescribeTable("should fail to create configurator given",
			func(ifaces []vmschema.Interface, networks []vmschema.Network) {
				_, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)

				Expect(err).To(HaveOccurred())
			},
			Entry("no interfaces",
				nil,
				[]vmschema.Network{{Name: "default", NetworkSource: vmschema.NetworkSource{Multus: &vmschema.MultusNetwork{}}}},
			),
			Entry("interface with no vhostuser binding method",
				[]vmschema.Interface{{Name: "default", InterfaceBindingMethod: vmschema.InterfaceBindingMethod{Bridge: &vmschema.InterfaceBridge{}}}},
				[]vmschema.Network{*vmschema.DefaultPodNetwork()},
			),
			Entry("interface with no vhostuser binding plugin",
				[]vmschema.Interface{{Name: "default", Binding: &vmschema.PluginBinding{Name: "no-vhostuser"}}},
				[]vmschema.Network{*vmschema.DefaultPodNetwork()},
			),
		)

		It("should fail given interface with invalid PCI address", func() {
			ifaces := []vmschema.Interface{{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"},
				PciAddress: "invalid-pci-address"}}
			networks := []vmschema.Network{*vmschema.DefaultPodNetwork()}

			testMutator, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)
			Expect(err).ToNot(HaveOccurred())

			_, err = testMutator.Mutate(&libvirtxml.Domain{})
			Expect(err).To(HaveOccurred())
		})

		DescribeTable("should add interface to domain spec given iface with",
			func(iface *vmschema.Interface, expectedDomainIface *libvirtxml.DomainInterface) {
				ifaces := []vmschema.Interface{*iface}
				networks := []vmschema.Network{*vmschema.DefaultPodNetwork()}

				testMutator, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)
				Expect(err).ToNot(HaveOccurred())

				mutatedDomain, err := testMutator.Mutate(&libvirtxml.Domain{})
				Expect(err).ToNot(HaveOccurred())
				Expect(mutatedDomain.Devices.Interfaces).To(Equal([]libvirtxml.DomainInterface{*expectedDomainIface}))
			},
			Entry("vhostuser binding plugin",
				&vmschema.Interface{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"}},
				&libvirtxml.DomainInterface{
					Alias: utils.NewUserDefinedAlias("default"),
					Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
				},
			),
			Entry("PCI address",
				&vmschema.Interface{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"},
					PciAddress: "0000:02:02.0"},
				&libvirtxml.DomainInterface{
					Alias:   utils.NewUserDefinedAlias("default"),
					Model:   &libvirtxml.DomainInterfaceModel{Type: "virtio"},
					Address: &libvirtxml.DomainAddress{PCI: pciAddr(0, 2, 2, 0)},
				},
			),
			Entry("MAC address",
				&vmschema.Interface{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"},
					MacAddress: "02:02:02:02:02:02"},
				&libvirtxml.DomainInterface{
					Alias: utils.NewUserDefinedAlias("default"),
					Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
					MAC:   &libvirtxml.DomainInterfaceMAC{Address: "02:02:02:02:02:02"},
				},
			),
			Entry("ACPI address",
				&vmschema.Interface{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"},
					ACPIIndex: 2},
				&libvirtxml.DomainInterface{
					Alias: utils.NewUserDefinedAlias("default"),
					Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
					ACPI:  &libvirtxml.DomainDeviceACPI{Index: uint(2)},
				},
			),
		)

		It("should not override other interfaces", func() {
			networks := []vmschema.Network{
				*vmschema.DefaultPodNetwork(),
				{Name: "secondary", NetworkSource: vmschema.NetworkSource{Multus: &vmschema.MultusNetwork{NetworkName: "sec"}}},
			}
			ifaces := []vmschema.Interface{
				{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"}},
				{Name: "secondary", InterfaceBindingMethod: vmschema.InterfaceBindingMethod{Bridge: &vmschema.InterfaceBridge{}}},
			}

			expectedDomainIface := &libvirtxml.DomainInterface{
				Alias: utils.NewUserDefinedAlias("default"),
				Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
			}

			testMutator, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)
			Expect(err).ToNot(HaveOccurred())

			existingIface := libvirtxml.DomainInterface{Alias: utils.NewUserDefinedAlias("existing-iface")}
			testDomain := &libvirtxml.Domain{
				Devices: &libvirtxml.DomainDeviceList{
					Interfaces: []libvirtxml.DomainInterface{existingIface},
				},
			}

			mutatedDomain, err := testMutator.Mutate(testDomain)
			Expect(err).ToNot(HaveOccurred())
			Expect(mutatedDomain.Devices.Interfaces).To(Equal([]libvirtxml.DomainInterface{existingIface, *expectedDomainIface}))
		})

		It("should set domain interface correctly when executed more than once", func() {
			networks := []vmschema.Network{*vmschema.DefaultPodNetwork()}
			ifaces := []vmschema.Interface{{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"}}}

			expectedDomainIface := &libvirtxml.DomainInterface{
				Alias: utils.NewUserDefinedAlias("default"),
				Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
			}

			testMutator, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)
			Expect(err).ToNot(HaveOccurred())

			testDomain := &libvirtxml.Domain{}

			mutatedDomain, err := testMutator.Mutate(testDomain)
			Expect(err).ToNot(HaveOccurred())
			Expect(mutatedDomain.Devices.Interfaces).To(Equal([]libvirtxml.DomainInterface{*expectedDomainIface}))

			mutatedAgain, err := testMutator.Mutate(mutatedDomain)
			Expect(err).ToNot(HaveOccurred())
			Expect(mutatedAgain).To(Equal(mutatedDomain))
		})

		It("should handle multiple vhostuser interfaces correctly", func() {
			networks := []vmschema.Network{
				*vmschema.DefaultPodNetwork(),
				{Name: "network1", NetworkSource: vmschema.NetworkSource{Multus: &vmschema.MultusNetwork{NetworkName: "net1"}}},
				{Name: "network2", NetworkSource: vmschema.NetworkSource{Multus: &vmschema.MultusNetwork{NetworkName: "net2"}}},
			}
			ifaces := []vmschema.Interface{
				{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"}},
				{Name: "net1", Binding: &vmschema.PluginBinding{Name: "vhostuser"}, MacAddress: "02:00:00:00:00:01"},
				{Name: "net2", Binding: &vmschema.PluginBinding{Name: "vhostuser"}, MacAddress: "02:00:00:00:00:02", PciAddress: "0000:03:00.0"},
			}

			expectedDomainIfaces := []libvirtxml.DomainInterface{
				{
					Alias: utils.NewUserDefinedAlias("default"),
					Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
				},
				{
					Alias: utils.NewUserDefinedAlias("net1"),
					Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
					MAC:   &libvirtxml.DomainInterfaceMAC{Address: "02:00:00:00:00:01"},
				},
				{
					Alias:   utils.NewUserDefinedAlias("net2"),
					Model:   &libvirtxml.DomainInterfaceModel{Type: "virtio"},
					MAC:     &libvirtxml.DomainInterfaceMAC{Address: "02:00:00:00:00:02"},
					Address: &libvirtxml.DomainAddress{PCI: pciAddr(0, 3, 0, 0)},
				},
			}

			testMutator, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)
			Expect(err).ToNot(HaveOccurred())

			mutatedDomain, err := testMutator.Mutate(&libvirtxml.Domain{})
			Expect(err).ToNot(HaveOccurred())
			Expect(mutatedDomain.Devices.Interfaces).To(HaveLen(3))
			Expect(mutatedDomain.Devices.Interfaces).To(Equal(expectedDomainIfaces))
		})

		It("should replace existing interface with same name", func() {
			networks := []vmschema.Network{*vmschema.DefaultPodNetwork()}
			ifaces := []vmschema.Interface{{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"}}}

			testMutator, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)
			Expect(err).ToNot(HaveOccurred())

			existingIface := libvirtxml.DomainInterface{
				Alias: utils.NewUserDefinedAlias("default"),
				Source: &libvirtxml.DomainInterfaceSource{
					Bridge: &libvirtxml.DomainInterfaceSourceBridge{Bridge: "br0"},
				},
			}
			testDomain := &libvirtxml.Domain{
				Devices: &libvirtxml.DomainDeviceList{
					Interfaces: []libvirtxml.DomainInterface{existingIface},
				},
			}

			expectedDomainIface := &libvirtxml.DomainInterface{
				Alias: utils.NewUserDefinedAlias("default"),
				Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
			}

			mutatedDomain, err := testMutator.Mutate(testDomain)
			Expect(err).ToNot(HaveOccurred())
			Expect(mutatedDomain.Devices.Interfaces).To(HaveLen(1))
			Expect(mutatedDomain.Devices.Interfaces[0]).To(Equal(*expectedDomainIface))
		})

		It("should handle mixed vhostuser and non-vhostuser interfaces", func() {
			networks := []vmschema.Network{
				*vmschema.DefaultPodNetwork(),
				{Name: "multus1", NetworkSource: vmschema.NetworkSource{Multus: &vmschema.MultusNetwork{NetworkName: "net1"}}},
				{Name: "multus2", NetworkSource: vmschema.NetworkSource{Multus: &vmschema.MultusNetwork{NetworkName: "net2"}}},
			}
			ifaces := []vmschema.Interface{
				{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"}},
				{Name: "multus1", InterfaceBindingMethod: vmschema.InterfaceBindingMethod{Bridge: &vmschema.InterfaceBridge{}}},
				{Name: "multus2", Binding: &vmschema.PluginBinding{Name: "vhostuser"}},
			}

			testMutator, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)
			Expect(err).ToNot(HaveOccurred())

			existingBridgeIface := libvirtxml.DomainInterface{
				Alias: utils.NewUserDefinedAlias("multus1"),
				Source: &libvirtxml.DomainInterfaceSource{
					Bridge: &libvirtxml.DomainInterfaceSourceBridge{Bridge: "br0"},
				},
			}
			testDomain := &libvirtxml.Domain{
				Devices: &libvirtxml.DomainDeviceList{
					Interfaces: []libvirtxml.DomainInterface{existingBridgeIface},
				},
			}

			mutatedDomain, err := testMutator.Mutate(testDomain)
			Expect(err).ToNot(HaveOccurred())
			Expect(mutatedDomain.Devices.Interfaces).To(HaveLen(3))

			// Bridge interface should remain untouched
			Expect(mutatedDomain.Devices.Interfaces[0].Source.Bridge).ToNot(BeNil())
			Expect(utils.AliasName(mutatedDomain.Devices.Interfaces[0].Alias)).To(Equal("multus1"))

			// Vhostuser interfaces should be added (no Source set yet)
			vhostuserIfaces := 0
			for _, iface := range mutatedDomain.Devices.Interfaces {
				if iface.Source == nil || iface.Source.VHostUser == nil {
					if iface.Model != nil && iface.Model.Type == "virtio" {
						vhostuserIfaces++
					}
				}
			}
			Expect(vhostuserIfaces).To(Equal(2))
		})

		It("should set memory backing to shared", func() {
			ifaces := []vmschema.Interface{{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"}}}
			networks := []vmschema.Network{*vmschema.DefaultPodNetwork()}

			testMutator, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)
			Expect(err).ToNot(HaveOccurred())

			mutatedDomain, err := testMutator.Mutate(&libvirtxml.Domain{})
			Expect(err).ToNot(HaveOccurred())
			Expect(mutatedDomain.MemoryBacking).ToNot(BeNil())
			Expect(mutatedDomain.MemoryBacking.MemoryAccess).ToNot(BeNil())
			Expect(mutatedDomain.MemoryBacking.MemoryAccess.Mode).To(Equal("shared"))
		})

		It("should set memory backing to shared even if it exists with a different mode", func() {
			ifaces := []vmschema.Interface{{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"}}}
			networks := []vmschema.Network{*vmschema.DefaultPodNetwork()}

			testMutator, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)
			Expect(err).ToNot(HaveOccurred())

			testDomain := &libvirtxml.Domain{
				MemoryBacking: &libvirtxml.DomainMemoryBacking{
					MemoryAccess: &libvirtxml.DomainMemoryAccess{Mode: "private"},
				},
			}

			mutatedDomain, err := testMutator.Mutate(testDomain)
			Expect(err).ToNot(HaveOccurred())
			Expect(mutatedDomain.MemoryBacking).ToNot(BeNil())
			Expect(mutatedDomain.MemoryBacking.MemoryAccess).ToNot(BeNil())
			Expect(mutatedDomain.MemoryBacking.MemoryAccess.Mode).To(Equal("shared"))
		})

		It("should set memory backing access when backing exists but access is nil", func() {
			ifaces := []vmschema.Interface{{Name: "default", Binding: &vmschema.PluginBinding{Name: "vhostuser"}}}
			networks := []vmschema.Network{*vmschema.DefaultPodNetwork()}

			testMutator, err := domain.NewVhostUserNetworkConfigurator(ifaces, networks)
			Expect(err).ToNot(HaveOccurred())

			testDomain := &libvirtxml.Domain{
				MemoryBacking: &libvirtxml.DomainMemoryBacking{},
			}

			mutatedDomain, err := testMutator.Mutate(testDomain)
			Expect(err).ToNot(HaveOccurred())
			Expect(mutatedDomain.MemoryBacking).ToNot(BeNil())
			Expect(mutatedDomain.MemoryBacking.MemoryAccess).ToNot(BeNil())
			Expect(mutatedDomain.MemoryBacking.MemoryAccess.Mode).To(Equal("shared"))
		})
	})
})
