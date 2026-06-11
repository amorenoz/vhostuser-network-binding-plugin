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
 * Copyright 2026 Red Hat, Inc.
 *
 */

// Package ovsdpdk implements [driver.DRADriver] for the OVS-DPDK DRA driver
// (https://github.com/amorenoz/dra-driver-ovsdpdk).
//
// It reads device metadata via a [metadata.DRAMetadataProvider] and extracts
// the vhost-user socket path from the driver-specific attribute.
package ovsdpdk

import (
	"fmt"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/klog/v2"

	"kubevirt.io/vhostuser-network-binding-plugin/pkg/dra/driver"
	"kubevirt.io/vhostuser-network-binding-plugin/pkg/dra/metadata"
)

const (
	// DriverName is the DRA driver name used by the OVS-DPDK DRA driver.
	DriverName = "ovsdpdk.k8snetworkplumbingwg.io"

	// VhostPathKey is the device attribute key that the OVS-DPDK DRA driver
	// uses to publish the vhost-user socket path.
	VhostPathKey = "vhost-user-path"
)

// OvsDpdkDriver implements [driver.DRADriver] for the OVS-DPDK DRA driver.
type OvsDpdkDriver struct {
	meta metadata.DRAMetadataProvider
}

// NewOvsDpdkDriver returns an OvsDpdkDriver backed by the given provider.
func NewOvsDpdkDriver(meta metadata.DRAMetadataProvider) *OvsDpdkDriver {
	return &OvsDpdkDriver{meta: meta}
}

// GetVhostMetadata implements [driver.DRADriver].  It reads the OVS-DPDK DRA
// driver metadata for the given ResourceClaimTemplate claim and request, and
// extracts the vhost-user socket path.
func (o *OvsDpdkDriver) GetVhostMetadata(claimName, requestName string) (driver.VhostMetadata, error) {
	dm, err := o.meta.Read(DriverName, claimName, requestName)
	if err != nil {
		return driver.VhostMetadata{}, fmt.Errorf("read OVS-DPDK DRA metadata for claim %q request %q: %w", claimName, requestName, err)
	}

	key := resourceapi.QualifiedName(VhostPathKey)
	var result *driver.VhostMetadata
	for _, req := range dm.Requests {
		for _, dev := range req.Devices {
			attr, ok := dev.Attributes[key]
			if !ok {
				continue
			}
			if attr.StringValue == nil {
				return driver.VhostMetadata{}, fmt.Errorf("attribute %q in claim %q request %q is not a string",
					VhostPathKey, claimName, requestName)
			}
			if result != nil {
				klog.Warningf("ovsdpdk: multiple devices carry attribute %q in claim %q request %q; using the first one",
					VhostPathKey, claimName, requestName)
				break
			}
			vhostPath := *attr.StringValue
			klog.Infof("ovsdpdk: DRA metadata resolved %q=%q for claim %q request %q",
				VhostPathKey, vhostPath, claimName, requestName)
			result = &driver.VhostMetadata{VhostPath: vhostPath}
		}
	}

	if result != nil {
		return *result, nil
	}
	return driver.VhostMetadata{}, fmt.Errorf("attribute %q not found in OVS-DPDK DRA metadata for claim %q request %q",
		VhostPathKey, claimName, requestName)
}
