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

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" // Register pprof HTTP handlers
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/lucas-clemente/quic-go"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	pb "github.com/anapaya/bwallocation/proto/bwreserver/v1"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/fatal"
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
)

func Cmd(pather command.Pather) *cobra.Command {
	var flags struct {
		dispatcher string
		sciond     string
		local      flag.UDPAddr
		pprof      flag.TCPAddr
		logLevel   string
	}
	cmd := &cobra.Command{
		Use:     "server",
		Short:   "Bandwidth Reservation Demo Server",
		Example: "  " + pather.CommandPath() + " server --local :9000",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			local, err := resolveLocal(flags.sciond, (*net.UDPAddr)(&flags.local))
			if err != nil {
				return err
			}
			if err := app.SetupLog(flags.logLevel); err != nil {
				return serrors.WrapStr("setting up logging", err)
			}

			dispatcher := flags.dispatcher
			v, ok := os.LookupEnv("SCION_DISPATCHER_SOCKET")
			if ok && dispatcher == reliable.DefaultDispPath {
				dispatcher = v
			}
			server, err := newServer(dispatcher, local)
			if err != nil {
				return err
			}

			ctx := app.WithSignal(context.Background(), os.Interrupt, syscall.SIGTERM)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			errs := make(chan error)
			if flags.pprof.Port != 0 {
				go func() {
					defer log.HandlePanic()
					if err := http.ListenAndServe(flags.pprof.String(), nil); err != nil {
						errs <- err
					}
				}()
			}
			go func() {
				defer log.HandlePanic()
				errs <- server.run(ctx)
			}()

			return <-errs
		},
	}
	flags.local = flag.UDPAddr{Port: 9000}
	cmd.Flags().StringVar(&flags.dispatcher, "dispatcher", reliable.DefaultDispPath,
		"dispatcher socket")
	cmd.Flags().Var(&flags.local, "local", "Address to listen on")
	cmd.Flags().StringVar(&flags.sciond, "sciond", sciond.DefaultAPIAddress, "SCION Daemon address")
	cmd.Flags().Var(&flags.pprof, "pprof", "Address to serve pprof")
	cmd.Flags().StringVar(&flags.logLevel, "log.level", "", app.LogLevelUsage)

	return cmd
}

func resolveLocal(sd string, srvAddr *net.UDPAddr) (*snet.UDPAddr, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	daemon, err := sciond.NewService(sd).Connect(ctx)
	if err != nil {
		return nil, serrors.WrapStr("connecting to SCION Daemon", err)
	}
	info, err := app.QueryASInfo(ctx, daemon)
	if err != nil {
		return nil, serrors.WrapStr("querying AS info", err)
	}
	if len(srvAddr.IP) == 0 {
		if srvAddr.IP, err = addrutil.DefaultLocalIP(ctx, daemon); err != nil {
			return nil, serrors.WrapStr("determine default local IP", err)
		}
	}
	return &snet.UDPAddr{
		IA:   info.IA,
		Host: srvAddr,
	}, nil
}

const (
	maxSessions       = 20
	sessionExpiration = 5 * time.Second
	reportInterval    = 1 * time.Second
	quicErrNoErr      = "Application error 0x100"
)

type server struct {
	pb.UnimplementedSessionServiceServer

	network           *snet.SCIONNetwork
	ctrlListener      net.Listener
	host              *net.UDPAddr
	tlsConfig         *tls.Config
	sessions          map[int]session
	statsChan         chan sessionStatsUpdate
	sessionClosedChan chan int
	mu                sync.Mutex
	wg                sync.WaitGroup
}

func newServer(disp string, l *snet.UDPAddr) (*server, error) {
	serverNet := &snet.SCIONNetwork{
		LocalIA: l.IA,
		Dispatcher: &snet.DefaultPacketDispatcherService{
			Dispatcher:  reliable.NewDispatcher(disp),
			SCMPHandler: ignoreSCMP{},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serverConn, err := serverNet.Listen(ctx, "udp", l.Host, addr.SvcNone)
	if err != nil {
		return nil, serrors.WrapStr("creating server connection", err)
	}
	tlsConfig, err := infraenv.GenerateTLSConfig()
	if err != nil {
		return nil, serrors.WrapStr("generating TLS config", err)
	}
	listener, err := quic.Listen(serverConn, tlsConfig, nil)
	if err != nil {
		return nil, serrors.WrapStr("listening QUIC/SCION", err)
	}

	return &server{
		network:           serverNet,
		ctrlListener:      squic.NewConnListener(listener),
		host:              l.Host,
		tlsConfig:         tlsConfig,
		sessions:          make(map[int]session, maxSessions),
		statsChan:         make(chan sessionStatsUpdate, maxSessions),
		sessionClosedChan: make(chan int, maxSessions),
	}, nil
}

func (s *server) run(ctx context.Context) error {
	grpcServer := grpc.NewServer()
	pb.RegisterSessionServiceServer(grpcServer, s)
	go func() {
		defer log.HandlePanic()
		fmt.Printf("Listening on %s for incoming connections.\n", s.host)
		if err := grpcServer.Serve(s.ctrlListener); err != nil {
			fatal.Fatal(err)
		}
	}()
	for {
		select {
		case update := <-s.statsChan:
			s.reportStats(update)
		case id := <-s.sessionClosedChan:
			s.mu.Lock()
			delete(s.sessions, id)
			s.mu.Unlock()
		case <-ctx.Done():
			s.mu.Lock()
			defer s.mu.Unlock()
			fmt.Println("interrupt received. closing all sessions...")
			for _, session := range s.sessions {
				session.close()
			}
			s.wg.Wait()
			fmt.Println("all sessions closed. terminating...")
			return nil
		}
	}
}

func (s *server) reportStats(stats sessionStatsUpdate) {
	receiveRate := float64(stats.bytesReceived*8) / stats.elapsed.Seconds() / 1e6
	fmt.Printf("session %d\tremote: %s\telapsed time: %s\treceive rate: %.2fmbps\n",
		stats.sessionID, stats.remote, stats.elapsed.Round(time.Second), receiveRate)
}

func (s *server) Stats(ctx context.Context,
	req *pb.StatsRequest) (*pb.StatsResponse, error) {

	id := int(req.SessionId)
	session, ok := s.sessions[id]
	if !ok {
		return nil, grpcstatus.Error(codes.Internal, "session id was not found")
	}

	e := session.stat()
	start, err := ptypes.TimestampProto(e.start)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, "server encountered internal error")
	}
	current, err := ptypes.TimestampProto(e.curr)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, "server encountered internal error")
	}
	return &pb.StatsResponse{
		Start:         start,
		Current:       current,
		BytesReceived: uint64(e.bytes),
	}, nil
}

func (s *server) OpenSession(ctx context.Context,
	request *pb.OpenSessionRequest) (*pb.OpenSessionResponse, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	srcIA := addr.IAInt(request.SrcIsdAs).IA()
	srcHost := net.IP(request.SrcHost)
	fmt.Printf("Session request received from %s,%s\n", srcIA, srcHost)

	if len(s.sessions) == maxSessions {
		return nil, grpcstatus.Error(codes.ResourceExhausted,
			"server cannot handle any more sessions")
	}

	protocol := request.Protocol
	if protocol == pb.Protocol_PROTOCOL_UNSPECIFIED {
		protocol = pb.Protocol_PROTOCOL_UDP
	}

	// Create new data conn.
	l := &net.UDPAddr{
		IP:   s.host.IP,
		Port: 0,
		Zone: s.host.Zone,
	}
	c, err := s.network.Listen(ctx, "udp", l, addr.SvcNone)
	if err != nil {
		return nil, grpcstatus.Error(codes.Unavailable, "server cannot open UDP socket")
	}
	localAddr := c.LocalAddr().(*net.UDPAddr)
	port := localAddr.Port
	if _, ok := s.sessions[port]; ok {
		// Lingering session for this port still exists. Cannot create a new one.
		c.Close()
		return nil, grpcstatus.Error(codes.Internal, "server encountered internal error")
	}
	var session session
	if protocol == pb.Protocol_PROTOCOL_UDP {
		session, err = s.createUDPSession(localAddr, c)
	} else {
		session, err = s.createQUICSession(localAddr, c)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating data session: %s\n", err)
		c.Close()
		return nil, grpcstatus.Error(codes.Unavailable, "server cannot create data session")
	}

	s.sessions[port] = session

	go func() {
		defer log.HandlePanic()
		defer s.wg.Done()

		s.wg.Add(1)
		session.serve()
	}()

	return &pb.OpenSessionResponse{
		SessionId: uint64(port),
		Port:      uint32(port),
		Protocol:  protocol,
	}, nil
}

func (s *server) createUDPSession(local *net.UDPAddr, c *snet.Conn) (session, error) {
	return &udpSession{
		sessionBase: sessionBase{
			id:                local.Port,
			local:             local,
			statsChan:         s.statsChan,
			sessionClosedChan: s.sessionClosedChan,
			closeChan:         make(chan struct{}),
		},
		conn: c,
	}, nil
}

func (s *server) createQUICSession(local *net.UDPAddr, c *snet.Conn) (session, error) {
	listener, err := quic.Listen(c, s.tlsConfig, nil)
	if err != nil {
		return nil, err
	}
	return &quicSession{
		sessionBase: sessionBase{
			id:                local.Port,
			local:             local,
			statsChan:         s.statsChan,
			sessionClosedChan: s.sessionClosedChan,
			closeChan:         make(chan struct{}),
		},
		listener: squic.NewConnListener(listener),
	}, nil
}

type sessionStatsUpdate struct {
	sessionID     int
	remote        net.Addr
	elapsed       time.Duration
	bytesReceived int
}

type session interface {
	serve()
	stat() *status
	close()
}

type status struct {
	start, curr time.Time
	bytes       int
}

type sessionBase struct {
	id    int
	local *net.UDPAddr

	statsChan         chan sessionStatsUpdate
	closeChan         chan struct{}
	sessionClosedChan chan int

	mu                 sync.RWMutex
	bytesReceivedTotal int
	bytesReceived      int
	start              time.Time
	lastReport         time.Time
}

func (s *sessionBase) stat() *status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &status{
		start: s.start,
		curr:  time.Now(),
		bytes: s.bytesReceivedTotal,
	}
}

func (s *sessionBase) collectStats(n int, remote net.Addr) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bytesReceived += n
	now := time.Now()
	elapsed := now.Sub(s.lastReport)
	if elapsed >= reportInterval {
		s.statsChan <- sessionStatsUpdate{
			sessionID:     s.id,
			remote:        remote,
			elapsed:       elapsed,
			bytesReceived: s.bytesReceived,
		}
		s.lastReport = now
		s.bytesReceived = 0
	}
	s.bytesReceivedTotal += n
}

func (s *sessionBase) close() {
	close(s.closeChan)
}

type udpSession struct {
	sessionBase
	conn *snet.Conn
}

func (s *udpSession) serve() {
	if s.conn == nil {
		panic("serve called before data conn was initialized")
	}
	defer s.conn.Close()

	fmt.Printf("session %d starting (UDP)...\n", s.id)

	localAddrLen := net.IPv6len
	if v4 := s.local.IP.To4(); v4 != nil {
		localAddrLen = net.IPv4len
	}
	s.bytesReceived = 0
	s.bytesReceivedTotal = 0
	s.start = time.Now()
	s.lastReport = time.Now()

	recvBuf := make([]byte, 9000)
	received := 0
	for {
		select {
		case <-s.closeChan:
			return
		default:
			recvBuf = recvBuf[:cap(recvBuf)]
			s.conn.SetReadDeadline(time.Now().Add(sessionExpiration))
			n, sender, err := s.conn.ReadFrom(recvBuf)
			if err != nil && serrors.IsTimeout(err) {
				fmt.Printf("session %d timed out. closing...\n", s.id)
				s.sessionClosedChan <- s.id
				return
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading from conn: %s\n", err)
				s.sessionClosedChan <- s.id
				return
			}
			// Account for SCION Header and throughput.
			received = n + headerLen(sender.(*snet.UDPAddr), localAddrLen)
			s.collectStats(received, sender)
		}
	}
}

type quicSession struct {
	sessionBase
	listener *squic.ConnListener
}

func (s *quicSession) serve() {
	if s.listener == nil {
		panic("serve called before QUIC listener was initialized")
	}
	defer s.listener.Close()
	ctx, cancel := context.WithTimeout(context.Background(), sessionExpiration)
	defer cancel()
	conn, err := s.listener.AcceptCtx(ctx)
	if err != nil && serrors.IsTimeout(err) {
		fmt.Printf("session %d timed out. closing...\n", s.id)
		s.sessionClosedChan <- s.id
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listening for QUIC data connections: %s\n", err)
		s.sessionClosedChan <- s.id
		return
	}

	fmt.Printf("session %d starting (QUIC)...\n", s.id)

	sender := conn.RemoteAddr().(*snet.UDPAddr)
	localAddrLen := net.IPv6len
	if v4 := s.local.IP.To4(); v4 != nil {
		localAddrLen = net.IPv4len
	}
	s.bytesReceived = 0
	s.lastReport = time.Now()

	recvBuf := make([]byte, 9000)
	received := 0
	for {
		select {
		case <-s.closeChan:
			return
		default:
			recvBuf = recvBuf[:cap(recvBuf)]
			conn.SetReadDeadline(time.Now().Add(sessionExpiration))
			n, err := conn.Read(recvBuf)
			if err != nil {
				if serrors.IsTimeout(err) || err.Error() == quicErrNoErr {
					fmt.Printf("session %d timed out. closing...\n", s.id)
					s.sessionClosedChan <- s.id
					return
				}
				fmt.Fprintf(os.Stderr, "error reading from conn: %s\n", err)
				s.sessionClosedChan <- s.id
				return
			}
			// Account for SCION Header and throughput.
			received = n + headerLen(sender, localAddrLen)
			s.collectStats(received, sender)
		}
	}
}

func headerLen(sender *snet.UDPAddr, localAddrLen int) int {
	senderAddrLen := net.IPv6len
	if v4 := sender.Host.IP.To4(); v4 != nil {
		senderAddrLen = net.IPv4len
	}
	return slayers.CmnHdrLen +
		2*addr.IABytes +
		senderAddrLen +
		localAddrLen +
		len(sender.Path.Raw) +
		8 // UDP/SCION header
}

type ignoreSCMP struct{}

func (ignoreSCMP) Handle(pkt *snet.Packet) error {
	// Always reattempt reads from the socket.
	return nil
}
