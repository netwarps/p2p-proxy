package endpoint

import (
	"context"
	"errors"
	"net"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	p2pproto "github.com/libp2p/go-libp2p/core/protocol"
	"go.uber.org/multierr"

	"github.com/diandianl/p2p-proxy/config"
	"github.com/diandianl/p2p-proxy/endpoint/balancer"
	"github.com/diandianl/p2p-proxy/log"
	"github.com/diandianl/p2p-proxy/p2p"
	"github.com/diandianl/p2p-proxy/protocol"
	"github.com/diandianl/p2p-proxy/relay"
)

func New(cfg *config.Config) (*Endpoint, error) {
	if err := cfg.Validate(false); err != nil {
		return nil, err
	}
	return &Endpoint{
		logger:   log.NewSubLogger("Endpoint"),
		cfg:      cfg,
		stopping: make(chan struct{}),
	}, nil
}

type Endpoint struct {
	logger log.Logger

	cfg *config.Config

	node host.Host

	listeners []protocol.Listener

	balancer balancer.Balancer

	stopping chan struct{}
}

func (e *Endpoint) Start(ctx context.Context) (err error) {

	logger := e.logger
	defer logger.Sync()

	logger.Info("Starting Endpoint")

	c := e.cfg

	if len(c.Endpoint.ProxyProtocols) == 0 {
		return errors.New("'Config.Endpoint.ProxyProtocols' can not be empty")
	}

	e.balancer, err = balancer.New(c.Endpoint.Balancer, e)
	if err != nil {
		return err
	}

	logger.Debugf("Endpoint using '%s' balancer", e.balancer.Name())

	e.node, err = p2p.NewHost(ctx, c)
	if err != nil {
		return err
	}

	for _, p := range c.Endpoint.ProxyProtocols {
		lsr, err := protocol.NewListener(protocol.Protocol(p.Protocol), p.Listen)
		if err != nil {
			return err
		}
		logger.Infof("Enable %s service, listen at: %s", lsr.Protocol(), p.Listen)
		e.listeners = append(e.listeners, lsr)

		go func() {
			err := e.startListener(ctx, lsr)
			if err != nil {
				e.logger.Errorf("start proxy listener [%s], ", lsr.Protocol(), err)
			}
		}()
	}

	<-ctx.Done()
	return e.Stop()
}

func (e *Endpoint) startListener(ctx context.Context, lsr protocol.Listener) error {
	for {
		conn, err := lsr.Accept()
		if err != nil {
			return e.errorTriggeredByStop(err)
		}
		go e.connHandler(ctx, lsr.Protocol(), conn)
	}
}

func (e *Endpoint) connHandler(ctx context.Context, p protocol.Protocol, conn net.Conn) {
	stream, err := e.newProxyStream(ctx, p, 3)
	// If an error happens, we write an error for response.
	if err != nil {
		if e.errorTriggeredByStop(err) != nil {
			e.logger.Warn("New stream ", err)
		}
		return
	}
	if err := relay.CloseAfterRelay(conn, stream); e.errorTriggeredByStop(err) != nil {
		e.logger.Warn("Relay failure: ", err)
	}
}

func (e *Endpoint) isStopping() bool {
	select {
	case <-e.stopping:
		return true
	default:
		return false
	}
}

func (e *Endpoint) errorTriggeredByStop(err error) error {
	select {
	case <-e.stopping:
		return nil
	default:
		return err
	}
}

func (e *Endpoint) newProxyStream(ctx context.Context, p protocol.Protocol, retry int) (network.Stream, error) {
	proxy, err := e.balancer.Next(p)
	if err != nil {
		if balancer.IsNewNotEnoughProxiesError(err) && retry > 0 {
			return e.newProxyStream(ctx, p, 0)
		}
		return nil, err
	}
	s, err := e.node.NewStream(ctx, proxy, p2pproto.ID(p))
	if err != nil {
		retry--
		if retry <= 0 {
			return nil, err
		}
		return e.newProxyStream(ctx, p, retry)
	}
	return s, nil
}

func (e *Endpoint) GetProxies(p protocol.Protocol) (ret []peer.ID) {
	self := e.node.ID()
	for _, id := range e.node.Peerstore().Peers() {
		if id != self {
			ret = append(ret, id)
		}
	}
	return
}

func (e *Endpoint) Stop() error {
	close(e.stopping)
	errs := make([]error, 0, len(e.listeners)+1)
	for _, lsr := range e.listeners {
		errs = append(errs, lsr.Close())
	}
	errs = append(errs, e.node.Close())
	return multierr.Combine(errs...)
}
