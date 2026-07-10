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

package main

import (
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"

	"kubevirt.io/kubevirt/pkg/hooks"
	hooksInfo "kubevirt.io/kubevirt/pkg/hooks/info"
	hooksV1alpha3 "kubevirt.io/kubevirt/pkg/hooks/v1alpha3"

	"kubevirt.io/client-go/log"

	draMetadata "kubevirt.io/vhostuser-network-binding-plugin/pkg/dra/metadata"
	"kubevirt.io/vhostuser-network-binding-plugin/pkg/dra/ovsdpdk"
	srv "kubevirt.io/vhostuser-network-binding-plugin/pkg/server"
)

const hookSocket = "vhostuser.sock"

func main() {
	draDriver := ovsdpdk.NewOvsDpdkDriver(draMetadata.DRAMetadata{})

	socketPath := filepath.Join(hooks.HookSocketsSharedDirectory, hookSocket)
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Log.Reason(err).Errorf("Failed to initialize socket on path: %s", socketPath)
		log.Log.Error("Check whether given directory exists and socket name is not already taken by other file")
		os.Exit(1)
	}
	defer func() {
		if err := os.Remove(socketPath); err != nil {
			log.Log.Reason(err).Errorf("Failed to remove socket %s", socketPath)
		}
	}()

	server := grpc.NewServer([]grpc.ServerOption{}...)
	hooksInfo.RegisterInfoServer(server, srv.InfoServer{Version: "v1alpha3"})

	shutdownChan := make(chan struct{})
	hooksV1alpha3.RegisterCallbacksServer(server, srv.NewV1alpha3Server(draDriver, shutdownChan))

	log.Log.Infof("Starting hook server exposing 'info' and '%s' services on socket %q", "v1alpha3", socketPath)
	srv.Serve(server, socket, shutdownChan)
}
