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

package ovsdpdk_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	resourceapi "k8s.io/api/resource/v1"
	drametadata "k8s.io/dynamic-resource-allocation/api/metadata"

	"kubevirt.io/vhostuser-network-binding-plugin/pkg/dra/ovsdpdk"
)

// mockProvider implements metadata.DRAMetadataProvider for testing.
type mockProvider struct {
	dm             *drametadata.DeviceMetadata
	err            error
	capturedDriver string
}

func (m *mockProvider) Read(driverName, _, _ string) (*drametadata.DeviceMetadata, error) {
	m.capturedDriver = driverName
	return m.dm, m.err
}

// deviceMetadataWithAttr builds a minimal DeviceMetadata containing a single
// device with the given attribute key/value pair.
func deviceMetadataWithAttr(key, value string) *drametadata.DeviceMetadata {
	return &drametadata.DeviceMetadata{
		Requests: []drametadata.DeviceMetadataRequest{
			{
				Name: "vhost-port",
				Devices: []drametadata.Device{
					{
						Driver: ovsdpdk.DriverName,
						Pool:   "node-0",
						Name:   "dev0",
						Attributes: map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{
							resourceapi.QualifiedName(key): {StringValue: new(value)},
						},
					},
				},
			},
		},
	}
}

var _ = Describe("OvsDpdkDriver", func() {
	const (
		claimName   = "net1"
		requestName = "vhost-port"
		vhostPath   = "/var/run/ovsdpdk/vhost-user/net1/vhost.sock"
	)

	Describe("GetVhostMetadata", func() {
		Context("when the provider returns metadata with the vhost-user-path attribute", func() {
			It("returns VhostMetadata with the correct path", func() {
				provider := &mockProvider{dm: deviceMetadataWithAttr(ovsdpdk.VhostPathKey, vhostPath)}
				d := ovsdpdk.NewOvsDpdkDriver(provider)

				result, err := d.GetVhostMetadata(claimName, requestName)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.VhostPath).To(Equal(vhostPath))
			})

			It("passes the OVS-DPDK driver name to the provider", func() {
				provider := &mockProvider{dm: deviceMetadataWithAttr(ovsdpdk.VhostPathKey, vhostPath)}
				d := ovsdpdk.NewOvsDpdkDriver(provider)

				_, err := d.GetVhostMetadata(claimName, requestName)
				Expect(err).ToNot(HaveOccurred())
				Expect(provider.capturedDriver).To(Equal(ovsdpdk.DriverName))
			})
		})

		Context("when the provider returns an error", func() {
			It("propagates the error", func() {
				provider := &mockProvider{err: fmt.Errorf("metadata not found")}
				d := ovsdpdk.NewOvsDpdkDriver(provider)

				_, err := d.GetVhostMetadata(claimName, requestName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("metadata not found"))
			})
		})

		Context("when the metadata contains no vhost-user-path attribute", func() {
			It("returns an error mentioning the attribute key and claim", func() {
				provider := &mockProvider{dm: deviceMetadataWithAttr("some-other-key", "some-value")}
				d := ovsdpdk.NewOvsDpdkDriver(provider)

				_, err := d.GetVhostMetadata(claimName, requestName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(ovsdpdk.VhostPathKey))
				Expect(err.Error()).To(ContainSubstring(claimName))
			})
		})

		Context("when the vhost-user-path attribute is not a string", func() {
			It("returns an error mentioning the attribute key", func() {
				intVal := int64(42)
				dm := &drametadata.DeviceMetadata{
					Requests: []drametadata.DeviceMetadataRequest{
						{
							Name: requestName,
							Devices: []drametadata.Device{
								{
									Attributes: map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{
										resourceapi.QualifiedName(ovsdpdk.VhostPathKey): {IntValue: &intVal},
									},
								},
							},
						},
					},
				}
				provider := &mockProvider{dm: dm}
				d := ovsdpdk.NewOvsDpdkDriver(provider)

				_, err := d.GetVhostMetadata(claimName, requestName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(ovsdpdk.VhostPathKey))
				Expect(err.Error()).To(ContainSubstring("not a string"))
			})
		})
	})
})
