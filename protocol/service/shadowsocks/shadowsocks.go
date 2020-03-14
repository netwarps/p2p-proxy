package shadowsocks

import (
	"context"
	"github.com/diandianl/p2p-proxy/log"
	"github.com/diandianl/p2p-proxy/protocol"
	"github.com/diandianl/p2p-proxy/relay"
	"github.com/shadowsocks/go-shadowsocks2/socks"
	"io"
	"net"
	"time"

	sscore "github.com/shadowsocks/go-shadowsocks2/core"
)

func init() {
	err := protocol.RegisterServiceFactory(protocol.Shadowsocks, "shadowsocks", New)
	if err != nil {
		panic(err)
	}
}

func New(logger log.Logger, cfg map[string]interface{}) (protocol.Service, error) {

	var (
		ciper    = "AES-128-CFB"
		password = "123456"
	)

	if c, ok := cfg["Ciper"]; ok {
		ciper = c.(string)
	}
	if p, ok := cfg["Password"]; ok {
		password = p.(string)
	}

	cip, err := sscore.PickCipher(ciper, nil, password)

	if err != nil {
		return nil, err
	}
	return &shadowsocksService{logger: logger, shadow: cip.StreamConn}, nil
}

type shadowsocksService struct {
	logger log.Logger

	shadow func(net.Conn) net.Conn

	listener net.Listener

	shuttingDown bool
}

func (_ *shadowsocksService) Protocol() protocol.Protocol {
	return protocol.Shadowsocks
}

func (s *shadowsocksService) Serve(ctx context.Context, l net.Listener) error {
	s.listener = l
	for {
		c, err := l.Accept()
		if err != nil {
			return s.errorTriggeredByShutdown(err)
		}
		go s.handleConn(c)
	}
}

func (s *shadowsocksService) handleConn(conn net.Conn) {

	defer conn.Close()

	logger := s.logger

	conn = s.shadow(conn)
	tgt, err := socks.ReadAddr(conn)
	if err != nil {
		if s.errorTriggeredByShutdown(err) != nil {
			logger.Warn("Read addr ", err)
		}
		return
	}

	rc, err := net.Dial("tcp", tgt.String())
	if err != nil {
		logger.Warnf("dial to target [%s] ", tgt, err)
		return
	}

	if rc, ok := rc.(*net.TCPConn); ok {
		err := rc.SetKeepAlive(true)
		if err != nil {
			logger.Warn("Set remote connection keepalive ", err)
		}
	}

	if err := relay.CloseAfterRelay(rc, conn); s.errorTriggeredByShutdown(err) != nil {
		logger.Warn("Relay failure ", err)
	}
}

func (s *shadowsocksService) errorTriggeredByShutdown(err error) error {
	if s.shuttingDown {
		return nil
	}
	return err
}

// relay copies between left and right bidirectionally. Returns number of
// bytes copied from right to left, from left to right, and any error occurred.
func relayx(left, right net.Conn) (int64, int64, error) {
	type res struct {
		N   int64
		Err error
	}
	ch := make(chan res)

	go func() {
		n, err := io.Copy(right, left)
		right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
		left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
		ch <- res{n, err}
	}()

	n, err := io.Copy(left, right)
	right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
	left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
	rs := <-ch

	if err == nil {
		err = rs.Err
	}
	return n, rs.N, err
}

func (s *shadowsocksService) Shutdown(ctx context.Context) error {
	s.shuttingDown = true
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
