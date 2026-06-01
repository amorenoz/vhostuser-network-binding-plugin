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

package callback_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	libvirtxml "libvirt.org/go/libvirtxml"

	"kubevirt.io/vhostuser-network-binding-plugin/pkg/callback"
	"kubevirt.io/vhostuser-network-binding-plugin/pkg/utils"
)

var _ = Describe("vhostuser hook callback handler", func() {
	Context("on define domain", func() {
		It("should fail given empty byte slice stream", func() {
			_, err := callback.OnDefineDomain([]byte{}, mutatorStub{})
			Expect(err).To(HaveOccurred())
		})

		It("should fail given invalid domain XML", func() {
			_, err := callback.OnDefineDomain([]byte("invalid-domain-xml"), mutatorStub{})
			Expect(err).To(HaveOccurred())
		})

		It("should fail when domain spec mutator fails", func() {
			domain := &libvirtxml.Domain{Name: "test"}
			domainXML, err := domain.Marshal()
			Expect(err).ToNot(HaveOccurred())

			expectedErr := fmt.Errorf("test error")
			domSpecMutator := mutatorStub{failMutate: expectedErr}

			_, err = callback.OnDefineDomain([]byte(domainXML), domSpecMutator)
			Expect(err).To(Equal(expectedErr))
		})

		It("given no-op mutator, domain spec should not change", func() {
			domain := &libvirtxml.Domain{Name: "test"}
			domainXML, err := domain.Marshal()
			Expect(err).ToNot(HaveOccurred())

			domSpecMutator := mutatorStub{domain: domain}

			result, err := callback.OnDefineDomain([]byte(domainXML), domSpecMutator)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(Equal(domainXML))
		})

		It("domain spec should mutate successfully", func() {
			domain := &libvirtxml.Domain{Name: "test"}
			domainXML, err := domain.Marshal()
			Expect(err).ToNot(HaveOccurred())

			mutatedDomain := &libvirtxml.Domain{
				Name: "test",
				Devices: &libvirtxml.DomainDeviceList{
					Interfaces: []libvirtxml.DomainInterface{
						{Alias: utils.NewUserDefinedAlias("test")},
					},
				},
			}
			domSpecMutator := mutatorStub{domain: mutatedDomain}

			mutatedDomainXML, err := mutatedDomain.Marshal()
			Expect(err).ToNot(HaveOccurred())

			result, err := callback.OnDefineDomain([]byte(domainXML), domSpecMutator)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(Equal(mutatedDomainXML))
		})
	})
})

type mutatorStub struct {
	domain     *libvirtxml.Domain
	failMutate error
}

func (s mutatorStub) Mutate(_ *libvirtxml.Domain) (*libvirtxml.Domain, error) {
	return s.domain, s.failMutate
}
