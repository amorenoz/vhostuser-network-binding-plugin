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

package utils

import (
	"strconv"
	"strings"

	hw "kubevirt.io/kubevirt/pkg/util/hardware"
	libvirtxml "libvirt.org/go/libvirtxml"
)

// User-defined aliases must have this prefix.
const userAliasPrefix = "ua-"

// NewUserDefinedAlias returns a DomainAlias automatically prepending
// the prefix.
func NewUserDefinedAlias(name string) *libvirtxml.DomainAlias {
	return &libvirtxml.DomainAlias{Name: userAliasPrefix + name}
}

// AliasName returns the logical name of an alias, stripping the prefix
// if present.
func AliasName(alias *libvirtxml.DomainAlias) string {
	if alias == nil {
		return ""
	}
	return strings.TrimPrefix(alias.Name, userAliasPrefix)
}

// NewPCIAddress parses a PCI address string in the format "domain:bus:slot.function"
// (e.g. "0000:02:02.0") into a libvirtxml.DomainAddressPCI.
func NewPCIAddress(addr string) (*libvirtxml.DomainAddressPCI, error) {
	partsStr, err := hw.ParsePciAddress(addr)
	if err != nil {
		return nil, err
	}
	domain, err := strconv.ParseUint(partsStr[0], 16, 32)
	if err != nil {
		return nil, err
	}
	bus, err := strconv.ParseUint(partsStr[1], 16, 32)
	if err != nil {
		return nil, err
	}
	slot, err := strconv.ParseUint(partsStr[2], 16, 32)
	if err != nil {
		return nil, err
	}
	function, err := strconv.ParseUint(partsStr[3], 16, 32)
	if err != nil {
		return nil, err
	}

	domainUint := uint(domain)
	busUint := uint(bus)
	slotUint := uint(slot)
	functionUint := uint(function)

	return &libvirtxml.DomainAddressPCI{
		Domain:   &domainUint,
		Bus:      &busUint,
		Slot:     &slotUint,
		Function: &functionUint,
	}, nil
}

// EnsureSharedMemoryBacking ensures the domain has shared memory backing.
func EnsureSharedMemoryBacking(domain *libvirtxml.Domain) {
	if domain.MemoryBacking == nil {
		domain.MemoryBacking = &libvirtxml.DomainMemoryBacking{
			MemoryAccess: &libvirtxml.DomainMemoryAccess{Mode: "shared"},
		}
		return
	}

	if domain.MemoryBacking.MemoryAccess == nil {
		domain.MemoryBacking.MemoryAccess = &libvirtxml.DomainMemoryAccess{Mode: "shared"}
		return
	}

	if domain.MemoryBacking.MemoryAccess.Mode != "shared" {
		domain.MemoryBacking.MemoryAccess.Mode = "shared"
	}
}
