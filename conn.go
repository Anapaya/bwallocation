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

package bwallocation

import (
	"net"

	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
)

// Conn is a wrapper for a snet.Conn that injects the reservation token
// fetched from the configured subscription.
type Conn struct {
	*snet.Conn
	Subscription *Subscription
}

func (c *Conn) Write(b []byte) (int, error) {
	return c.WriteTo(b, c.Conn.RemoteAddr())
}

func (c *Conn) WriteTo(p []byte, remote net.Addr) (int, error) {
	token, err := c.Subscription.Token()
	if err != nil {
		return 0, err
	}
	r, ok := remote.(*snet.UDPAddr)
	if !ok {
		return 0, serrors.New("invalid address type", "network", remote.Network())
	}
	a := r.Copy()
	a.Path = *token.Path
	return c.Conn.WriteTo(p, a)
}
