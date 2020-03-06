// Copyright (c) 2020 Doc.ai and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"

	"github.com/spiffe/spire/proto/spire/common"

	federation "github.com/networkservicemesh/networkservicemesh/applications/federation-server/api"
	"github.com/networkservicemesh/networkservicemesh/sdk/monitor"
)

func New() federation.RegistrationServer {
	rv := &federationServer{
		Server: monitor.NewServer(&eventFactory{
			factoryName: "Bundles",
		}),
	}
	go rv.Serve()
	return rv
}

type federationServer struct {
	monitor.Server
}

type bundleEntity struct {
	*common.Bundle
}

func (b *bundleEntity) GetId() string {
	return b.TrustDomainId
}

func (f *federationServer) CreateFederatedBundle(ctx context.Context, bundle *common.Bundle) (*common.Empty, error) {
	f.Update(ctx, &bundleEntity{bundle})
	return &common.Empty{}, nil
}

func (f *federationServer) ListFederatedBundles(_ *common.Empty, stream federation.Registration_ListFederatedBundlesServer) error {
	f.MonitorEntities(stream)
	return nil
}