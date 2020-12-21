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
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	bpb "github.com/anapaya/bwallocation/proto/bw_allocation/v1"
	"github.com/anapaya/bwallocation/reservation"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/slayers/path/scion"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/spath"
	"github.com/scionproto/scion/go/lib/util"
	"github.com/scionproto/scion/go/pkg/grpc"
	"google.golang.org/grpc/status"
)

const (
	defaultFetchInterval = time.Second
	defaultFetchLeadTime = 20 * time.Second
	defaultMaxExpiry     = 5 * time.Minute
)

// Manager manages bandwidth reservation subscriptions that can be used to
// to send traffic with reserved bandwidth.
type Manager struct {
	// Service is the address that exposes the bandwidth allocation service.
	// Usually, this should be the SCION daemon address.
	Service net.Addr
	// Dialer is the gRPC dialer that is used to connect to the service.
	Dialer grpc.Dialer
	// Logger is used for logging. If nil, nothing is logged.
	Logger log.Logger

	// FetchInterval indicates the minimum interval between successive
	// reservation renewal. If zero, the default interval is used.
	FetchInterval time.Duration
	// FetchLeadTime indicates the minimum remaining token validity before the
	// reservation is renewed. If zero, default lead time is used.
	FetchLeadTime time.Duration
	// MaxExpiry indicates the maximum reservation validity period that is
	// requested. If zero, the default max expiry is used.
	MaxExpiry time.Duration
}

// Subscribe subscribes for a bandwidth reservation. The subscription
// automatically renews the reservation token and must be closed after it is
// no longer required.
func (m *Manager) Subscribe(ctx context.Context, info SubscriptionInfo) (*Subscription, error) {
	id := uint32(rand.Intn(math.MaxInt32))

	interval, lead, maxExpiry := m.FetchInterval, m.FetchLeadTime, m.MaxExpiry
	if interval == 0 {
		interval = defaultFetchInterval
	}
	if lead == 0 {
		lead = defaultFetchLeadTime
	}
	if maxExpiry == 0 {
		maxExpiry = defaultMaxExpiry
	}

	sub := &Subscription{
		info:          info,
		id:            id,
		fetchInterval: interval,
		fetchLeadTime: lead,
		maxExpiry:     maxExpiry,
		service:       m.Service,
		dialer:        m.Dialer,
		logger:        log.SafeNewLogger(m.Logger, "id", fmt.Sprintf("%x", id)),
	}
	if err := sub.initial(ctx); err != nil {
		return nil, err
	}
	ctx, sub.cancel = context.WithCancel(context.Background())
	go func() {
		defer log.HandlePanic()
		sub.run(ctx)
	}()
	return sub, nil
}

// Token is the reservation token used to send traffic. Client code usually does
// not need to interact with the token directly.
type Token struct {
	Source      *snet.UDPAddr
	Destination *snet.UDPAddr
	Expiration  time.Time
	Path        *spath.Path
}

// SubscriptionInfo is the information for a bandwidth reservation subscription.
type SubscriptionInfo struct {
	// Source is the local address of the connection bandwidth should be
	// reserved on.
	Source *snet.UDPAddr
	// Destination is the remote address of the connection bandwidth should be
	// reserved on.
	Destination *snet.UDPAddr
	// Path is the original SCION path on which a bandwidth reservation should
	// be established.
	Path *spath.Path
	// Bandwidth is the bandwidth the reservation should have. It must be a
	// multiple of 1mbps, i.e., divisible by 1e6.
	Bandwidth Bandwidth
	// Expiration is the desired expiration time. For the zero value, the
	// Manager continuously renews the reservation. In case the bandwidth
	// allocation service is not willing to handout a reservation until expiry.
	// The manager continuously renews reservations until the expiration time is
	// reached.
	Expiration time.Time
}

// Subscription is a bandwidth reservation subscription that is periodically
// updated with a token.
type Subscription struct {
	info   SubscriptionInfo
	id     uint32
	cancel func()

	service       net.Addr
	dialer        grpc.Dialer
	logger        log.Logger
	fetchInterval time.Duration
	fetchLeadTime time.Duration
	maxExpiry     time.Duration

	currentMtx sync.RWMutex
	current    Token
	closed     bool
}

// Token returns the currently active token. The contents must not be modified.
func (s *Subscription) Token() (Token, error) {
	s.currentMtx.RLock()
	defer s.currentMtx.RUnlock()

	if s.closed {
		return Token{}, serrors.New("subscription closed")
	}
	return s.current, nil
}

// Close closes the subscription. It must be called after the subscription
// is no longer required.
func (s *Subscription) Close() {
	s.cancel()
}

func (s *Subscription) initial(ctx context.Context) error {
	if s.info.Path.Type != scion.PathType {
		return serrors.New("unsupported path type", "type", s.info.Path.Type)
	}
	if err := (&scion.Decoded{}).DecodeFromBytes(s.info.Path.Raw); err != nil {
		return serrors.WrapStr("invalid path", err)
	}
	return s.fetchToken(ctx)

}

func (s *Subscription) run(ctx context.Context) {
	timer := time.NewTimer(s.nextFetch())
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			s.currentMtx.Lock()
			defer s.currentMtx.Unlock()
			s.closed = true
			log.SafeDebug(s.logger, "Subscription canceled", "err", ctx.Err())
			return
		case now := <-timer.C:
			if !s.info.Expiration.After(now) && !s.info.Expiration.IsZero() {
				s.currentMtx.Lock()
				defer s.currentMtx.Unlock()
				s.closed = true
				log.SafeDebug(s.logger, "Subscription expired")
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			if err := s.fetchToken(ctx); err != nil {
				log.SafeDebug(s.logger, "Failed to fetch reservation", "err", err)
			}
			cancel()
			timer.Reset(s.nextFetch())
		}
	}
}

func (s *Subscription) fetchToken(ctx context.Context) error {
	cc, err := s.dialer.Dial(ctx, s.service)
	if err != nil {
		return err
	}
	defer cc.Close()

	expiry := s.info.Expiration
	if s.info.Expiration.IsZero() || time.Until(s.info.Expiration) > s.maxExpiry {
		expiry = time.Now().Add(s.maxExpiry).Truncate(time.Second)
	} else {
		expiry = expiry.Add(time.Second).Truncate(time.Second)
	}

	client := bpb.NewBandwithAllocationServiceClient(cc)
	rep, err := client.Reserve(ctx, &bpb.ReserveRequest{
		SrcHost:        canonicalIP(s.info.Source.Host.IP),
		DstIsdAs:       uint64(s.info.Destination.IA.IAInt()),
		DstHost:        canonicalIP(s.info.Destination.Host.IP),
		ExpirationTime: expiry.Unix(),
		Bandwidth:      uint64(s.info.Bandwidth),
		Id:             s.id,
		Path:           s.info.Path.Raw,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && len(st.Details()) > 0 {
			return serrors.WithCtx(err, "details", st.Details())
		}
		return err
	}

	var decoded scion.Decoded
	if err := decoded.DecodeFromBytes(s.info.Path.Raw); err != nil {
		return err
	}
	p := &reservation.Path{
		MetaHdr: reservation.MetaHdr{
			ExpTime:   util.TimeToSecs(expiry),
			ID:        s.id,
			MAC:       rep.Mac,
			RateLimit: uint16(s.info.Bandwidth / 1e6),
		},
		Decoded: decoded,
	}
	buf := make([]byte, p.Len())
	if err := p.SerializeTo(buf); err != nil {
		return err
	}

	s.currentMtx.Lock()
	defer s.currentMtx.Unlock()
	s.current = Token{
		Source:      s.info.Source,
		Destination: s.info.Destination,
		Path: &spath.Path{
			Raw:  buf,
			Type: p.Type(),
		},
		Expiration: util.SecsToTime(p.ExpTime),
	}
	return nil
}

// nextFetch returns:
// - time until subscription expiration, if the token covers it fully.
// - time until lead time before token expiration, if token expiration is more
//   than the lead time in the future
// - fetch interval if expiration is closer than the lead time.
func (s *Subscription) nextFetch() time.Duration {
	s.currentMtx.RLock()
	defer s.currentMtx.RUnlock()

	tokenExpiration := s.current.Expiration
	// Wait until expiration if token covers the full remaining time.
	if !s.info.Expiration.IsZero() && !tokenExpiration.Before(s.info.Expiration) {
		return time.Until(s.info.Expiration)
	}
	if time.Until(tokenExpiration) > s.fetchLeadTime {
		return time.Until(tokenExpiration.Add(-s.fetchLeadTime))
	}
	return s.fetchInterval
}

func canonicalIP(ip net.IP) net.IP {
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	return ip
}
