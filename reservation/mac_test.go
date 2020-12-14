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

package reservation_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/anapaya/bwallocation/reservation"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/scrypto"
	"github.com/scionproto/scion/go/lib/slayers"
	"github.com/scionproto/scion/go/lib/slayers/path/scion"
	"github.com/scionproto/scion/go/lib/xtest"
)

func TestMAC(t *testing.T) {
	spkt := &slayers.SCION{
		Version:      0,
		TrafficClass: 0xb8,
		FlowID:       0xdead,
		NextHdr:      common.L4UDP,
		PathType:     scion.PathType,
		DstAddrType:  slayers.T16Ip,
		DstAddrLen:   slayers.AddrLen16,
		SrcAddrType:  slayers.T4Ip,
		SrcAddrLen:   slayers.AddrLen4,
		DstIA:        xtest.MustParseIA("1-ff00:0:111"),
		SrcIA:        xtest.MustParseIA("2-ff00:0:222"),
	}
	ip6Addr := &net.IPAddr{IP: net.ParseIP("2001:db8::68")}
	ip4Addr := &net.IPAddr{IP: net.ParseIP("10.0.0.100")}
	spkt.SetDstAddr(ip6Addr)
	spkt.SetSrcAddr(ip4Addr)

	p := &reservation.Path{}
	p.DecodeFromBytes(rawPath)

	h, err := scrypto.InitMac(make([]byte, 16))
	assert.NoError(t, err)

	p.MAC = reservation.MAC(h, spkt, p)
	assert.NoError(t, reservation.VerifyMAC(h, spkt, p))
}
