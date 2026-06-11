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

// Package metadata provides DRA device metadata reading functionality.
package metadata

import (
	drametadata "k8s.io/dynamic-resource-allocation/api/metadata"
	"k8s.io/dynamic-resource-allocation/devicemetadata"
)

// DRAMetadataProvider reads DRA device metadata (KEP-5304) for a
// ResourceClaimTemplate-generated claim.
type DRAMetadataProvider interface {
	// Read returns the DeviceMetadata written by driverName for the given
	// ResourceClaimTemplate claim and request.
	Read(driverName, claimName, requestName string) (*drametadata.DeviceMetadata, error)
}

// DRAMetadata is the production implementation of [DRAMetadataProvider].
// It delegates directly to the DRA library.
type DRAMetadata struct{}

func (DRAMetadata) Read(driverName, claimName, requestName string) (*drametadata.DeviceMetadata, error) {
	return devicemetadata.ReadResourceClaimTemplateMetadataWithDriverName(driverName, claimName, requestName)
}
