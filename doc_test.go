// Copyright 2020 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bwallocation_test

import (
	"context"
	"net"
	"time"

	"github.com/anapaya/bwallocation"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/spath"
	"github.com/scionproto/scion/go/pkg/grpc"
)

func Example() {
	var (
		// conn is the underlying connection that is used to send traffic.
		conn *snet.Conn
		// deamonAddr is the TCP address where the SCION daemon is exposed.
		daemonAddr net.Addr
		// bandwidth is the bandwidth that should be reserved.
		bandwidth bwallocation.Bandwidth
		// expiration is the time until the Manager shall renew the reservation.
		expiration time.Time
		// remote is the address of the remote application.
		remote *snet.UDPAddr
		// local is the address of this application.
		local *snet.UDPAddr
		// path is the raw path that is used for creating the reservation.
		path *spath.Path
	)

	// Initialize the manager with the local SCION daemon as the service. The
	// SCION daemon proxies the requests to the correct control service
	// instance.
	manager := bwallocation.Manager{
		Service: daemonAddr,
		Dialer:  grpc.SimpleDialer{},
	}

	// Create a subscription to get the reservation token.
	sub, err := manager.Subscribe(context.Background(), bwallocation.SubscriptionInfo{
		Bandwidth:   bandwidth,
		Destination: remote,
		Source:      local,
		Expiration:  expiration,
		Path:        path,
	})
	if err != nil {
		// Handle error.
	}
	// Close the subscription as soon as it is no longer needed.
	defer sub.Close()

	// Wrap the conn such that it uses the reservation token.
	rsvConn := &bwallocation.Conn{
		Conn:         conn,
		Subscription: sub,
	}
	rsvConn.Write([]byte("hello with reserved bandwidth"))
}
