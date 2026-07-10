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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	vmschema "kubevirt.io/api/core/v1"

	"kubevirt.io/client-go/log"

	hooksInfo "kubevirt.io/kubevirt/pkg/hooks/info"
	hooksV1alpha3 "kubevirt.io/kubevirt/pkg/hooks/v1alpha3"

	"kubevirt.io/vhostuser-network-binding-plugin/pkg/callback"
	"kubevirt.io/vhostuser-network-binding-plugin/pkg/domain"
	"kubevirt.io/vhostuser-network-binding-plugin/pkg/dra/driver"
)

type InfoServer struct {
	Version string
}

func (s InfoServer) Info(_ context.Context, _ *hooksInfo.InfoParams) (*hooksInfo.InfoResult, error) {
	return &hooksInfo.InfoResult{
		Name: "network-vhostuser-binding",
		Versions: []string{
			s.Version,
		},
		HookPoints: []*hooksInfo.HookPoint{
			{
				Name:     hooksInfo.OnDefineDomainHookPointName,
				Priority: 0,
			},
			{
				Name:     hooksInfo.ShutdownHookPointName,
				Priority: 0,
			},
		},
	}, nil
}

// V1alpha3Server implements the KubeVirt hooks v1alpha3 gRPC server.
type V1alpha3Server struct {
	draDriver driver.DRADriver
	Done      chan struct{}
}

// NewV1alpha3Server returns a V1alpha3Server backed by the given DRADriver.
func NewV1alpha3Server(draDriver driver.DRADriver, done chan struct{}) V1alpha3Server {
	return V1alpha3Server{draDriver: draDriver, Done: done}
}

func (s V1alpha3Server) OnDefineDomain(_ context.Context, params *hooksV1alpha3.OnDefineDomainParams) (*hooksV1alpha3.OnDefineDomainResult, error) {
	vmi := &vmschema.VirtualMachineInstance{}
	if err := json.Unmarshal(params.GetVmi(), vmi); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VMI: %v", err)
	}

	vhostuserConfigurator, err := domain.NewVhostUserNetworkConfigurator(vmi, s.draDriver)
	if err != nil {
		return nil, fmt.Errorf("failed to create vhostuser configurator: %v", err)
	}

	newDomainXML, err := callback.OnDefineDomain(params.GetDomainXML(), vhostuserConfigurator)
	if err != nil {
		return nil, err
	}

	return &hooksV1alpha3.OnDefineDomainResult{
		DomainXML: newDomainXML,
	}, nil
}

func (s V1alpha3Server) PreCloudInitIso(_ context.Context, params *hooksV1alpha3.PreCloudInitIsoParams) (*hooksV1alpha3.PreCloudInitIsoResult, error) {
	return &hooksV1alpha3.PreCloudInitIsoResult{
		CloudInitData: params.GetCloudInitData(),
	}, nil
}

func (s V1alpha3Server) Shutdown(_ context.Context, _ *hooksV1alpha3.ShutdownParams) (*hooksV1alpha3.ShutdownResult, error) {
	log.Log.Info("shutdown vhostuser network binding")
	s.Done <- struct{}{}
	return &hooksV1alpha3.ShutdownResult{}, nil
}

func waitForShutdown(server *grpc.Server, errChan <-chan error, shutdownChan <-chan struct{}) {
	signalStopChan := make(chan os.Signal, 1)
	signal.Notify(signalStopChan, os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	var err error
	select {
	case s := <-signalStopChan:
		log.Log.Infof("received signal: %s", s.String())
	case err = <-errChan:
		log.Log.Reason(err).Error("grpc server failed")
	case <-shutdownChan:
		log.Log.Info("received shutdown, exiting")
	}
	if err == nil {
		server.GracefulStop()
	}
}

func Serve(server *grpc.Server, socket net.Listener, shutdownChan <-chan struct{}) {
	errChan := make(chan error)
	go func() {
		errChan <- server.Serve(socket)
	}()
	waitForShutdown(server, errChan, shutdownChan)
}
