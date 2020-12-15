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

package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"

	"github.com/anapaya/bwallocation"
	"github.com/anapaya/bwallocation/cmd/bwreserver/client/view"
	"github.com/anapaya/bwallocation/cmd/bwreserver/telemetry"
	pb "github.com/anapaya/bwallocation/proto/bwreserver/v1"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/infra/infraenv"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/slayers"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/lib/snet/addrutil"
	"github.com/scionproto/scion/go/lib/snet/squic"
	"github.com/scionproto/scion/go/lib/sock/reliable"
	"github.com/scionproto/scion/go/pkg/app"
	"github.com/scionproto/scion/go/pkg/app/flag"
	"github.com/scionproto/scion/go/pkg/command"
	libgrpc "github.com/scionproto/scion/go/pkg/grpc"
)

const (
	pollingInterval = 1 * time.Second
)

func Cmd(pather command.Pather) *cobra.Command {
	var flags struct {
		reservation string
		rate        string
		quic        bool
		pldSize     int
		local       net.IP
		sciond      string
		duration    int
		interactive bool
		sequence    string
		pprof       flag.TCPAddr
		logLevel    string
		text        bool
	}
	cmd := &cobra.Command{
		Use:   "client <server-addr>",
		Short: "Bandwidth Reservation Demo Client",
		Example: fmt.Sprintf(`  %[1]s client 1-ff00:0:112,127.0.0.1:9000 -s 5mbps -d 10
  %[1]s client 1-ff00:0:112,127.0.0.1:9000 -s 10mbps -r 5mbps
  %[1]s client 1-ff00:0:112,127.0.0.1:9000 --text
`, pather.CommandPath()),
		Args: cobra.ExactArgs(1),
		Long: fmt.Sprintf(`Bandwidth Reservation Demo Client

The client supports two protocols UDP/SCION and QUIC/SCION. By default UDP/SCION
is used. QUIC/SCION can be enabled with the appropriate flag.

%s`, app.SequenceHelp),
		RunE: func(cmd *cobra.Command, args []string) error {
			remote, err := snet.ParseUDPAddr(args[0])
			if err != nil {
				return serrors.WrapStr("parsing remote", err)
			}

			var rate bwallocation.Bandwidth
			if err := rate.UnmarshalText([]byte(flags.rate)); err != nil {
				return serrors.WrapStr("parsing sending-rate", err)
			}

			var reservation bwallocation.Bandwidth
			switch flags.reservation {
			case "none":
			case "sending-rate":
				reservation = rate
			default:
				if err := reservation.UnmarshalText([]byte(flags.reservation)); err != nil {
					return serrors.WrapStr("parsing reservation", err)
				}
			}

			if err := app.SetupLog(flags.logLevel); err != nil {
				return serrors.WrapStr("setting up logging", err)
			}

			cmd.SilenceUsage = true

			var v *view.View
			if !flags.text {
				v, err = view.New()
				if err != nil {
					return serrors.WrapStr("setting up view", err)
				}
				defer v.Close()
			}

			client := &client{
				remote:      remote,
				view:        v,
				rate:        rate,
				reservation: reservation,
				payloadSize: flags.pldSize,
				duration:    time.Duration(flags.duration) * time.Second,
				quic:        flags.quic,
				interactive: flags.interactive,
				sequence:    flags.sequence,
				text:        flags.text,
			}
			ctx := app.WithSignal(context.Background(), os.Interrupt, syscall.SIGTERM)
			if err := client.init(ctx, flags.sciond, flags.local); err != nil {
				return serrors.WrapStr("initializing client", err)
			}

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			errs := make(chan error, 3)
			if flags.pprof.Port != 0 {
				go func() {
					defer log.HandlePanic()
					if err := http.ListenAndServe(flags.pprof.String(), nil); err != nil {
						errs <- err
					}
				}()
			}

			var wg sync.WaitGroup
			if !flags.text {
				wg.Add(1)
				go func() {
					defer log.HandlePanic()
					defer wg.Done()
					errs <- client.view.Run(ctx)
				}()
			}

			go func() {
				defer log.HandlePanic()
				errs <- client.run(ctx)
			}()

			if flags.text {
				return <-errs
			}
			if err := <-errs; err != nil {
				return err
			}
			wg.Wait()
			return nil
		},
	}
	cmd.Flags().StringVarP(&flags.rate, "sending-rate", "s", "1mbps",
		"rate the client attempts to send at")
	cmd.Flags().StringVarP(&flags.reservation, "reservation", "r", "sending-rate", ""+
		"bandwidth to reserve. Setting this lower than sending rate simulates malicious behavior."+
		"\nsupported values:\n"+
		"  <bandwidth>:  Reserve the specified bandwidth\n"+
		"  sending-rate: Use same value as sending-rate\n"+
		"  none:         Do not reserve any bandwidth\n",
	)
	cmd.Flags().BoolVar(&flags.quic, "quic", false,
		"use QUIC when sending data. If not specified, UDP is used.")
	cmd.Flags().IPVar(&flags.local, "local", nil, "IP address to listen on")
	cmd.Flags().StringVar(&flags.sciond, "sciond", sciond.DefaultAPIAddress, "SCION Daemon address")
	cmd.Flags().IntVarP(&flags.duration, "duration", "d", 0, ""+
		"duration of the data transmission in seconds.\n"+
		"0 or negative values will keep the data transmission going indefinitely.",
	)
	cmd.Flags().BoolVarP(&flags.interactive, "interactive", "i", false, "interactive mode")
	cmd.Flags().StringVar(&flags.sequence, "sequence", "", app.SequenceUsage)
	cmd.Flags().Var(&flags.pprof, "pprof", "Address to serve pprof")
	cmd.Flags().StringVar(&flags.logLevel, "log.level", "", app.LogLevelUsage)
	cmd.Flags().BoolVar(&flags.text, "text", false,
		"use simple text mode. If not specified, the CLI widgets output is used.")
	cmd.Flags().IntVar(&flags.pldSize, "payload", 1280, "payload size in bytes")

	return cmd
}

type client struct {
	local          *snet.UDPAddr
	remote         *snet.UDPAddr
	id             uint64
	scionNet       *snet.SCIONNetwork
	tlsConfig      *tls.Config
	grpcClientConn *grpc.ClientConn
	daemon         net.Addr

	rate        bwallocation.Bandwidth
	reservation bwallocation.Bandwidth
	payloadSize int
	duration    time.Duration
	quic        bool
	interactive bool
	sequence    string

	text bool
	view *view.View

	tsClient            *telemetry.TimeSeries
	clientSeriesBuilder *view.SeriesBuilder
	tsServer            *telemetry.TimeSeries
	serverSeriesBuilder *view.SeriesBuilder
}

func (c *client) init(ctx context.Context, sd string, localIP net.IP) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	c.tsClient = telemetry.NewTimeSeries(time.Now())
	c.tsServer = telemetry.NewTimeSeries(time.Now())

	if !c.text {
		c.clientSeriesBuilder = &view.SeriesBuilder{
			Graph:      view.GraphUpdaterFunc(c.view.UpdateClientGraph),
			WindowSize: 30,
		}
		c.serverSeriesBuilder = &view.SeriesBuilder{
			Graph:      view.GraphUpdaterFunc(c.view.UpdateServerGraph),
			WindowSize: 30,
		}
		c.view.UpdateGauge(0)
	}

	daemonAddr, err := net.ResolveTCPAddr("tcp", sd)
	if err != nil {
		return serrors.WrapStr("resolving SCION Daemon address", err)
	}
	c.daemon = daemonAddr

	daemon, err := sciond.NewService(sd).Connect(ctx)
	if err != nil {
		return serrors.WrapStr("connecting to SCION Daemon", err)
	}
	info, err := app.QueryASInfo(ctx, daemon)
	if err != nil {
		return serrors.WrapStr("querying AS info", err)
	}
	path, err := app.ChoosePath(ctx, daemon, c.remote.IA, c.interactive, false, c.sequence)
	if err != nil {
		return serrors.WrapStr("choosing path", err)
	}
	c.remote.Path = path.Path()
	c.remote.NextHop = path.UnderlayNextHop()
	c.printf("Using path: %s\n", path)

	l := localIP
	if l == nil {
		target := c.remote.Host.IP
		if c.remote.NextHop != nil {
			target = c.remote.NextHop.IP
		}
		if l, err = addrutil.ResolveLocal(target); err != nil {
			return serrors.WrapStr("resolving local address", err)

		}
	}
	c.local = &snet.UDPAddr{
		IA:   info.IA,
		Host: &net.UDPAddr{IP: l},
	}

	c.scionNet = &snet.SCIONNetwork{
		LocalIA: info.IA,
		Dispatcher: &snet.DefaultPacketDispatcherService{
			Dispatcher:  reliable.NewDispatcher(reliable.DefaultDispPath),
			SCMPHandler: ignoreSCMP{},
		},
	}
	// Let the dispatcher decide on the port for the client connection.
	clientAddr := &net.UDPAddr{
		IP:   c.local.Host.IP,
		Zone: c.local.Host.Zone,
	}
	clientConn, err := c.scionNet.Listen(ctx, "udp", clientAddr, addr.SvcNone)
	if err != nil {
		return serrors.WrapStr("creating client connection", err)
	}

	// Setup QUIC stack and gRPC client conn
	c.tlsConfig, err = infraenv.GenerateTLSConfig()
	if err != nil {
		return serrors.WrapStr("generating TLS config", err)
	}
	dialer := &squic.ConnDialer{
		Conn:      clientConn,
		TLSConfig: c.tlsConfig,
	}
	dialF := func(context.Context, string) (net.Conn, error) {
		return dialer.Dial(ctx, c.remote)
	}
	c.grpcClientConn, err = grpc.DialContext(ctx, c.remote.String(),
		grpc.WithInsecure(),
		grpc.WithContextDialer(dialF),
		grpc.WithBlock())
	if err != nil {
		return serrors.WrapStr("establishing gRPC connection to remote", err)
	}
	return nil
}

func (c *client) run(ctx context.Context) error {
	conn, cancel, err := c.setupDataSession(ctx)
	if err != nil {
		return serrors.WrapStr("establishing new data session", err)
	}
	defer cancel()
	defer c.grpcClientConn.Close()
	c.println("Data session established.")

	senderDoneChan := make(chan error)

	c.println("local endpoint:         ", conn.LocalAddr())
	c.println("remote control endpoint:", c.remote)
	c.println("remote data endpoint:   ", conn.RemoteAddr())

	sender := &sender{
		dataConn:    conn,
		remote:      c.remote,
		rate:        c.rate,
		reservation: c.reservation,
		payloadSize: c.payloadSize,
		duration:    c.duration,
		ratelimit:   !c.quic,
		text:        c.text,
		view:        c.view,
		tsClient:    c.tsClient,
	}

	fetchCtx, cancelFetching := context.WithCancel(ctx)
	go func() {
		defer log.HandlePanic()
		defer conn.Close()
		defer cancelFetching()
		senderDoneChan <- sender.send(ctx)
	}()

	go func() {
		defer log.HandlePanic()
		c.fetchStatsFromRemote(fetchCtx)
	}()

	uiUpdateTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-uiUpdateTicker.C:
			c.reportStats()
		case err := <-senderDoneChan:
			c.println("data transmission done")
			return err
		}
	}
}

func (c *client) println(a ...interface{}) {
	if c.text {
		fmt.Println(a...)
	} else {
		c.view.Print(fmt.Sprintln(a...))
	}
}

func (c *client) printf(format string, a ...interface{}) {
	if c.text {
		fmt.Printf(format, a...)
	} else {
		c.view.Print(fmt.Sprintf(format, a...))
	}
}

func (c *client) fetchStatsFromRemote(ctx context.Context) {
	client := pb.NewSessionServiceClient(c.grpcClientConn)
	request := &pb.StatsRequest{
		SessionId: uint64(c.id),
	}

	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			newCtx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			response, err := client.Stats(newCtx, request)
			if err != nil {
				select {
				case <-ctx.Done():
				default:
					c.println("Failed to fetch stats from server:", err)
				}
				continue
			}
			// Ignore errors, we can silently discard them
			c.tsServer.Add(
				telemetry.Entry{
					Timestamp: time.Now(),
					Value:     float64(response.BytesReceived * 8),
				},
			)
		}
	}
}

// setupDataSession sets up a new data session with the server and returns a net.Conn object that
// can be used to transmit the data. If bandwidth needs to be reserved, setupDataSession returns a
// cancel function that can be used to close the bandwidth reservation subscription. Otherwise,
// the cancel function is a no-op (but safe to be called).
func (c *client) setupDataSession(ctx context.Context) (net.Conn, func(), error) {
	grpcClient := pb.NewSessionServiceClient(c.grpcClientConn)
	protocol, desc := pb.Protocol_PROTOCOL_UDP, "udp"
	if c.quic {
		protocol, desc = pb.Protocol_PROTOCOL_QUIC, "quic"
	}
	c.printf("Establishing %s data session to %s\n", desc, c.remote)
	request := &pb.OpenSessionRequest{
		SrcIsdAs: uint64(c.local.IA.IAInt()),
		SrcHost:  canonicalIP(c.local.Host.IP),
		Protocol: protocol,
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	response, err := grpcClient.OpenSession(ctx, request)
	if err != nil {
		return nil, func() {}, serrors.WrapStr("requesting new data session", err)
	}
	if protocol != response.Protocol {
		return nil, func() {}, serrors.New("transport protocol mismatch", "expected", protocol,
			"actual", response.Protocol)
	}

	dataRemote := c.remote.Copy()
	dataRemote.Host = &net.UDPAddr{
		IP:   c.remote.Host.IP,
		Port: int(response.Port),
	}
	c.id = response.SessionId

	udpConn, err := c.scionNet.Dial(ctx, "udp", c.local.Host, dataRemote, addr.SvcNone)
	if err != nil {
		return nil, func() {}, serrors.WrapStr("dialing UDP connection", err)
	}
	conn := net.Conn(udpConn)
	closeF := func() {}
	if !c.bestEffort() {
		if conn, closeF, err = c.withReservation(ctx, udpConn); err != nil {
			udpConn.Close()
			return nil, nil, serrors.WrapStr("establishing reservation", err)
		}
	}
	switch protocol {
	case pb.Protocol_PROTOCOL_UDP:
		return conn, closeF, nil
	case pb.Protocol_PROTOCOL_QUIC:
		dialer := &squic.ConnDialer{
			Conn:      conn.(net.PacketConn),
			TLSConfig: c.tlsConfig,
		}
		quicConn, err := dialer.Dial(ctx, dataRemote)
		if err != nil {
			udpConn.Close()
			return nil, nil, serrors.WrapStr("dialing QUIC connection", err)
		}
		return quicConn, closeF, nil
	default:
		udpConn.Close()
		return nil, nil, serrors.New("invalid protocol", "protocol", protocol)
	}
}

func (c *client) withReservation(ctx context.Context, conn *snet.Conn) (net.Conn, func(), error) {
	manager := bwallocation.Manager{
		Service: c.daemon,
		Dialer:  libgrpc.SimpleDialer{},
		Logger:  log.Root(),
	}
	var expiration time.Time
	if c.duration > 0 {
		expiration = time.Now().Add(c.duration + 2*time.Second)
	}
	sub, err := manager.Subscribe(ctx, bwallocation.SubscriptionInfo{
		Bandwidth:   c.reservation,
		Destination: c.remote,
		Source:      c.local,
		Expiration:  expiration,
		Path:        &c.remote.Path,
	})
	if err != nil {
		return nil, nil, err
	}
	return &bwallocation.Conn{
		Conn:         conn,
		Subscription: sub,
	}, sub.Close, nil
}

func (c *client) reportStats() {
	// Latency of 1 second (data is 1 second in the past) to improve the chance that we have
	// readings available.
	windowStart := time.Now().Add(-2 * time.Second)
	windowEnd := time.Now().Add(-time.Second)
	clientRate := telemetry.PerSecond(c.tsClient, windowStart, windowEnd)
	serverRate := telemetry.PerSecond(c.tsServer, windowStart, windowEnd)

	clientRateMbps := clientRate / 1e6
	serverRateMbps := serverRate / 1e6

	if c.text {
		toZero := func(v float64) float64 {
			if math.IsNaN(v) {
				return 0
			}
			return v
		}
		fmt.Printf("sending rate:   %.2f mbps (client)\n", toZero(clientRateMbps))
		fmt.Printf("receiving rate: %.2f mbps (server)\n", toZero(serverRateMbps))
	} else {
		c.view.UpdateGauge(int(clientRate))
		c.clientSeriesBuilder.Update(clientRateMbps)
		c.serverSeriesBuilder.Update(serverRateMbps)
	}
}

func (c *client) bestEffort() bool {
	return c.reservation == 0
}

type sender struct {
	dataConn    net.Conn
	remote      *snet.UDPAddr
	rate        bwallocation.Bandwidth
	reservation bwallocation.Bandwidth
	payloadSize int
	duration    time.Duration
	ratelimit   bool
	text        bool
	view        *view.View

	tsClient *telemetry.TimeSeries
}

func (s *sender) send(ctx context.Context) error {
	if s.dataConn == nil {
		panic("send called before data session was initialized")
	}
	defer s.dataConn.Close()

	// TODO(shitz): this doesn't take into account the QUIC header.
	pktLen := calcPktLen(s.payloadSize, s.dataConn.LocalAddr(), s.remote)
	pps := int(s.rate) / (pktLen * 8)
	limiter := ratelimit.New(int(pps))

	pld := make([]byte, s.payloadSize)
	rand.Read(pld)

	bytesSent := 0
	start := time.Now()
	end := start.Add(s.duration)
	infinite := !end.After(start)

	if s.reservation != 0 {
		s.printf("reserved bandwidth:      %s\n", s.reservation)
	} else {
		s.println("reserved bandwidth:      none")
	}
	s.printf("desired sending rate:    %s pps: %d (at %dB + %dB = %dB packets)\n",
		s.rate, pps, s.payloadSize, pktLen-s.payloadSize, pktLen)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var t1 time.Time
			if s.ratelimit {
				t1 = limiter.Take()
			} else {
				t1 = time.Now()
			}
			if !infinite && t1.After(end) {
				return nil
			}
			_, err := s.dataConn.Write(pld)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to write packet")
				return err
			}
			// We want to measure the throughput including the SCION header.
			bytesSent += pktLen
			// Ignore errors, we can silently discard them
			s.tsClient.Add(telemetry.Entry{Timestamp: time.Now(), Value: float64(bytesSent * 8)})
		}
	}
}

func (s *sender) println(a ...interface{}) {
	if s.text {
		fmt.Println(a...)
	} else {
		s.view.Print(fmt.Sprintln(a...))
	}
}

func (s *sender) printf(format string, a ...interface{}) {
	if s.text {
		fmt.Printf(format, a...)
	} else {
		s.view.Print(fmt.Sprintf(format, a...))
	}
}

func calcPktLen(pldLen int, local net.Addr, remote net.Addr) int {
	localAddrLen, remoteAddrLen := net.IPv6len, net.IPv6len
	if v4 := local.(*net.UDPAddr).IP.To4(); v4 != nil {
		localAddrLen = net.IPv4len
	}
	remoteSUDP := remote.(*snet.UDPAddr)
	if v4 := remoteSUDP.Host.IP.To4(); v4 != nil {
		remoteAddrLen = net.IPv4len
	}
	return slayers.CmnHdrLen +
		2*addr.IABytes +
		remoteAddrLen +
		localAddrLen +
		len(remoteSUDP.Path.Raw) +
		8 + // UDP/SCION header
		pldLen
}

func canonicalIP(ip net.IP) net.IP {
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	return ip
}

type ignoreSCMP struct{}

func (ignoreSCMP) Handle(pkt *snet.Packet) error {
	// Always reattempt reads from the socket.
	return nil
}
