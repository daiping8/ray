// Copyright 2025 The Ray Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// grpcMaxMessageSize caps both outgoing and incoming gRPC messages handled by
// the dashboard agent, aligned with the C++ dashboard agent limit of 16 MiB.
const grpcMaxMessageSize = 16 << 20 // 16 MiB

// grpcServer wraps the dashboard agent gRPC server and its service registration.
type grpcServer struct {
	server   *grpc.Server
	listener net.Listener
}

// newGRPCServer builds the gRPC server, registers services from every module,
// and binds the configured listen address.
func newGRPCServer(cfg Config, mods []Module) (*grpcServer, error) {
	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(grpcMaxMessageSize),
		grpc.MaxSendMsgSize(grpcMaxMessageSize),
	)
	for _, m := range mods {
		if err := m.RegisterGRPC(s); err != nil {
			return nil, fmt.Errorf("register gRPC services: %w", err)
		}
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.NodeIP, cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("listen on gRPC addr: %w", err)
	}
	return &grpcServer{server: s, listener: lis}, nil
}

// serve accepts connections on the bound listener until the server is stopped.
func (s *grpcServer) serve() error {
	return s.server.Serve(s.listener)
}

// stop gracefully stops the gRPC server.
func (s *grpcServer) stop() {
	s.server.GracefulStop()
}
