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

package domain

import (
	"fmt"

	vmschema "kubevirt.io/api/core/v1"
	libvirtxml "libvirt.org/go/libvirtxml"

	"kubevirt.io/vhostuser-network-binding-plugin/pkg/utils"
)

type VhostUserNetworkConfigurator struct {
	vhostIfaces []*vmschema.Interface
}

const (
	// VhostUserPluginName vhost-user binding plugin name should be registered to Kubevirt through Kubevirt CR
	VhostUserPluginName = "vhostuser"
)

func NewVhostUserNetworkConfigurator(ifaces []vmschema.Interface, networks []vmschema.Network) (*VhostUserNetworkConfigurator, error) {

	vhostIfaces := make([]*vmschema.Interface, 0)
	for _, iface := range ifaces {
		if iface.Binding != nil && iface.Binding.Name == VhostUserPluginName {
			vhostIfaces = append(vhostIfaces, &iface)
		}
	}

	if len(vhostIfaces) == 0 {
		return nil, fmt.Errorf("no vhost interfaces found")
	}

	return &VhostUserNetworkConfigurator{
		vhostIfaces: vhostIfaces,
	}, nil
}

func (p VhostUserNetworkConfigurator) Mutate(domain *libvirtxml.Domain) (*libvirtxml.Domain, error) {
	if domain.Devices == nil {
		domain.Devices = &libvirtxml.DomainDeviceList{}
	}

	for _, vhostIface := range p.vhostIfaces {
		generatedIface, err := p.generateDomainInterface(vhostIface)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to generate domain interface spec for iface: %v", vhostIface.Name, err)
		}

		if iface := lookupIfaceByAliasName(domain.Devices.Interfaces, vhostIface.Name); iface != nil {
			*iface = *generatedIface
		} else {
			domain.Devices.Interfaces = append(domain.Devices.Interfaces, *generatedIface)
		}
	}

	return domain, nil
}

func (p VhostUserNetworkConfigurator) generateDomainInterface(iface *vmschema.Interface) (*libvirtxml.DomainInterface, error) {
	domIface := &libvirtxml.DomainInterface{
		Alias: utils.NewUserDefinedAlias(iface.Name),
		Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
		// TODO: Add vhostuser source
	}

	if iface.PciAddress != "" {
		pciAddr, err := utils.NewPCIAddress(iface.PciAddress)
		if err != nil {
			return nil, err
		}
		domIface.Address = &libvirtxml.DomainAddress{PCI: pciAddr}
	}

	if iface.MacAddress != "" {
		domIface.MAC = &libvirtxml.DomainInterfaceMAC{Address: iface.MacAddress}
	}

	if iface.ACPIIndex > 0 {
		domIface.ACPI = &libvirtxml.DomainDeviceACPI{Index: uint(iface.ACPIIndex)}
	}

	return domIface, nil
}

func lookupIfaceByAliasName(ifaces []libvirtxml.DomainInterface, name string) *libvirtxml.DomainInterface {
	for i, iface := range ifaces {
		if utils.AliasName(iface.Alias) == name {
			return &ifaces[i]
		}
	}
	return nil
}
