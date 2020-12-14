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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anapaya/bwallocation"
	bpb "github.com/anapaya/bwallocation/proto/bw_allocation/v1"
	mock_bw_allocation "github.com/anapaya/bwallocation/proto/bw_allocation/v1/mock_bw_allocation"
	"github.com/anapaya/bwallocation/reservation"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/slayers/path"
	"github.com/scionproto/scion/go/lib/slayers/path/scion"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/spath"
	"github.com/scionproto/scion/go/lib/util"
	"github.com/scionproto/scion/go/lib/xtest"
)

func TestManagerSubscribe(t *testing.T) {
	mustAddr := func(t *testing.T, s string) *snet.UDPAddr {
		var a snet.UDPAddr
		err := a.Set(s)
		require.NoError(t, err)
		return &a
	}
	newPath := func(t *testing.T) *spath.Path {
		decoded := scion.Decoded{
			Base: scion.Base{
				PathMeta: scion.MetaHdr{
					SegLen: [3]uint8{2, 2},
				},
				NumHops: 4,
				NumINF:  2,
			},
			InfoFields: []*path.InfoField{{}, {}},
			HopFields:  []*path.HopField{{}, {}, {}, {}},
		}
		raw := make([]byte, decoded.Len())
		err := decoded.SerializeTo(raw)
		require.NoError(t, err)

		return &spath.Path{
			Type: scion.PathType,
			Raw:  raw,
		}
	}
	t.Run("first success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := mock_bw_allocation.NewMockBandwithAllocationServiceServer(ctrl)
		srv.EXPECT().Reserve(gomock.Any(), gomock.Any()).Return(
			&bpb.ReserveResponse{
				Mac: []byte{1, 2, 3, 4, 5, 6},
			}, nil,
		)

		reg := xtest.NewGRPCService()
		bpb.RegisterBandwithAllocationServiceServer(reg.Server(), srv)
		reg.Start(t)

		manager := bwallocation.Manager{
			Service: addr.SvcCS,
			Dialer:  reg,
		}
		sub, err := manager.Subscribe(context.Background(), bwallocation.SubscriptionInfo{
			Source:      mustAddr(t, "1-ff00:0:111,127.0.0.1"),
			Destination: mustAddr(t, "1-ff00:0:112,127.0.0.2"),
			Bandwidth:   10e6,
			Expiration:  time.Now().Add(time.Minute),
			Path:        newPath(t),
		})
		require.NoError(t, err)
		token, err := sub.Token()
		require.NoError(t, err)
		assert.Equal(t, mustAddr(t, "1-ff00:0:111,127.0.0.1"), token.Source)
		assert.Equal(t, mustAddr(t, "1-ff00:0:112,127.0.0.2"), token.Destination)
		assert.InDelta(t, time.Until(token.Expiration).Seconds(), 60, 2)

		var p reservation.Path
		require.NoError(t, p.DecodeFromBytes(token.Path.Raw))
		assert.Equal(t, util.TimeToSecs(token.Expiration), p.ExpTime)
		assert.NotZero(t, p.ID)
		assert.Equal(t, uint16(10), p.RateLimit)
		assert.Equal(t, []byte{1, 2, 3, 4, 5, 6}, p.MAC)
	})
	t.Run("error after close", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := mock_bw_allocation.NewMockBandwithAllocationServiceServer(ctrl)
		srv.EXPECT().Reserve(gomock.Any(), gomock.Any()).Return(
			&bpb.ReserveResponse{
				Mac: []byte{1, 2, 3, 4, 5, 6},
			}, nil,
		)

		reg := xtest.NewGRPCService()
		bpb.RegisterBandwithAllocationServiceServer(reg.Server(), srv)
		reg.Start(t)

		manager := bwallocation.Manager{
			Service: addr.SvcCS,
			Dialer:  reg,
		}
		sub, err := manager.Subscribe(context.Background(), bwallocation.SubscriptionInfo{
			Source:      mustAddr(t, "1-ff00:0:111,127.0.0.1"),
			Destination: mustAddr(t, "1-ff00:0:112,127.0.0.2"),
			Bandwidth:   10e6,
			Expiration:  time.Now().Add(time.Minute),
			Path:        newPath(t),
		})
		require.NoError(t, err)
		_, err = sub.Token()
		require.NoError(t, err)

		sub.Close()
		time.Sleep(50 * time.Millisecond)

		_, err = sub.Token()
		assert.Error(t, err)
	})
	t.Run("closed after expiry", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := mock_bw_allocation.NewMockBandwithAllocationServiceServer(ctrl)
		srv.EXPECT().Reserve(gomock.Any(), gomock.Any()).Return(
			&bpb.ReserveResponse{
				Mac: []byte{1, 2, 3, 4, 5, 6},
			}, nil,
		)

		reg := xtest.NewGRPCService()
		bpb.RegisterBandwithAllocationServiceServer(reg.Server(), srv)
		reg.Start(t)

		manager := bwallocation.Manager{
			Service: addr.SvcCS,
			Dialer:  reg,
		}
		sub, err := manager.Subscribe(context.Background(), bwallocation.SubscriptionInfo{
			Source:      mustAddr(t, "1-ff00:0:111,127.0.0.1"),
			Destination: mustAddr(t, "1-ff00:0:112,127.0.0.2"),
			Bandwidth:   10e6,
			Expiration:  time.Now(),
			Path:        newPath(t),
		})
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		_, err = sub.Token()
		assert.Error(t, err)
	})
	t.Run("renews", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := mock_bw_allocation.NewMockBandwithAllocationServiceServer(ctrl)
		srv.EXPECT().Reserve(gomock.Any(), gomock.Any()).Return(
			&bpb.ReserveResponse{
				Mac: []byte{1, 2, 3, 4, 5, 6},
			}, nil,
		).MinTimes(2).MaxTimes(15)

		reg := xtest.NewGRPCService()
		bpb.RegisterBandwithAllocationServiceServer(reg.Server(), srv)
		reg.Start(t)

		manager := bwallocation.Manager{
			Service:       addr.SvcCS,
			Dialer:        reg,
			MaxExpiry:     time.Minute,
			FetchLeadTime: time.Minute,
			FetchInterval: 500 * time.Millisecond,
		}
		sub, err := manager.Subscribe(context.Background(), bwallocation.SubscriptionInfo{
			Source:      mustAddr(t, "1-ff00:0:111,127.0.0.1"),
			Destination: mustAddr(t, "1-ff00:0:112,127.0.0.2"),
			Bandwidth:   10e6,
			Expiration:  time.Now().Add(2 * time.Minute),
			Path:        newPath(t),
		})
		require.NoError(t, err)

		// We sleep here to give the manager time to renew. As the MinTimes
		// above are set to 2, this test fails if the manager does not issue a
		// renewal. The manager constantly renews because the lead time is equal
		// to the max expiry.
		time.Sleep(time.Second)

		token, err := sub.Token()
		require.NoError(t, err)
		assert.Equal(t, mustAddr(t, "1-ff00:0:111,127.0.0.1"), token.Source)
		assert.Equal(t, mustAddr(t, "1-ff00:0:112,127.0.0.2"), token.Destination)
		assert.InDelta(t, time.Until(token.Expiration).Seconds(), 60, 2)

		var p reservation.Path
		require.NoError(t, p.DecodeFromBytes(token.Path.Raw))
		assert.Equal(t, util.TimeToSecs(token.Expiration), p.ExpTime)
		assert.NotZero(t, p.ID)
		assert.Equal(t, uint16(10), p.RateLimit)
		assert.Equal(t, []byte{1, 2, 3, 4, 5, 6}, p.MAC)
	})
	t.Run("honor max", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := mock_bw_allocation.NewMockBandwithAllocationServiceServer(ctrl)
		srv.EXPECT().Reserve(gomock.Any(), gomock.Any()).Return(
			&bpb.ReserveResponse{
				Mac: []byte{1, 2, 3, 4, 5, 6},
			}, nil,
		)

		reg := xtest.NewGRPCService()
		bpb.RegisterBandwithAllocationServiceServer(reg.Server(), srv)
		reg.Start(t)

		manager := bwallocation.Manager{
			Service:   addr.SvcCS,
			Dialer:    reg,
			MaxExpiry: time.Minute,
		}
		sub, err := manager.Subscribe(context.Background(), bwallocation.SubscriptionInfo{
			Source:      mustAddr(t, "1-ff00:0:111,127.0.0.1"),
			Destination: mustAddr(t, "1-ff00:0:112,127.0.0.2"),
			Bandwidth:   10e6,
			Expiration:  time.Now().Add(5 * time.Minute),
			Path:        newPath(t),
		})
		require.NoError(t, err)

		token, err := sub.Token()
		require.NoError(t, err)
		assert.Equal(t, mustAddr(t, "1-ff00:0:112,127.0.0.2"), token.Destination)
		assert.True(t, time.Until(token.Expiration) <= time.Minute)
	})

}
