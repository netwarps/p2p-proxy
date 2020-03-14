package endpoint

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/diandianl/p2p-proxy/config"
	"github.com/diandianl/p2p-proxy/endpoint/balancer"
	"github.com/diandianl/p2p-proxy/log"
	"github.com/diandianl/p2p-proxy/p2p"
	"github.com/diandianl/p2p-proxy/protocol"
	"github.com/diandianl/p2p-proxy/relay"

	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	p2pproto "github.com/libp2p/go-libp2p-core/protocol"
	discovery2 "github.com/libp2p/go-libp2p-discovery"
	"go.uber.org/multierr"
)

type Endpoint interface {
	Start(ctx context.Context) error

	Stop() error
}

func New(cfg *config.Config) (Endpoint, error) {
	if err := cfg.Validate(false); err != nil {
		return nil, err
	}
	return &endpoint{
		logger:   log.NewSubLogger("endpoint"),
		cfg:      cfg,
		proxies:  make(map[peer.ID]struct{}),
		stopping: make(chan struct{}),
	}, nil
}

type endpoint struct {
	logger log.Logger

	cfg *config.Config

	node host.Host

	discoverer discovery.Discoverer

	listeners []protocol.Listener

	balancer balancer.Balancer

	sync.Mutex
	proxies map[peer.ID]struct{}

	stopping chan struct{}
}

func (e *endpoint) Start(ctx context.Context) (err error) {

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

	e.node, e.discoverer, err = p2p.NewHostAndDiscovererAndBootstrap(ctx, c)
	if err != nil {
		return err
	}

	go e.syncProxies(ctx)

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

func (e *endpoint) startListener(ctx context.Context, lsr protocol.Listener) error {
	for {
		conn, err := lsr.Accept()
		if err != nil {
			return e.errorTriggeredByStop(err)
		}
		go e.connHandler(ctx, lsr.Protocol(), conn)
	}
}

func (e *endpoint) connHandler(ctx context.Context, p protocol.Protocol, conn net.Conn) {
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

func (e *endpoint) isStopping() bool {
	select {
	case <-e.stopping:
		return true
	default:
		return false
	}
}

func (e *endpoint) errorTriggeredByStop(err error) error {
	select {
	case <-e.stopping:
		return nil
	default:
		return err
	}
}

func (e *endpoint) newProxyStream(ctx context.Context, p protocol.Protocol, retry int) (network.Stream, error) {
	proxy, err := e.balancer.Next(p)
	if err != nil {
		if balancer.IsNewNotEnoughProxiesError(err) && retry > 0 {
			proxies, err := e.DiscoveryProxies(ctx)
			if err != nil {
				return nil, err
			}
			e.UpdateProxies(proxies)
			return e.newProxyStream(ctx, p, 0)
		}
		return nil, err
	}
	s, err := e.node.NewStream(ctx, proxy, p2pproto.ID(p))
	if err != nil {
		e.DeleteProxy(proxy)
		retry--
		if retry <= 0 {
			return nil, err
		}
		return e.newProxyStream(ctx, p, retry)
	}
	return s, nil
}

func (e *endpoint) syncProxies(ctx context.Context) {
	ticker := time.NewTicker(e.cfg.Endpoint.ServiceDiscoveryInterval)
	for {
		if proxies, err := e.DiscoveryProxies(ctx); err != nil {
			e.logger.Error(err)
		} else {
			e.UpdateProxies(proxies)
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (e *endpoint) UpdateProxies(proxies []peer.ID) {
	e.Lock()
	defer e.Unlock()
	for _, p := range proxies {
		e.proxies[p] = struct{}{}
	}
}

func (e *endpoint) GetProxies(p protocol.Protocol) []peer.ID {
	if len(e.proxies) == 0 {
		return nil
	}
	e.Lock()
	defer e.Unlock()
	proxies := make([]peer.ID, 0, len(e.proxies))
	for p, _ := range e.proxies {
		proxies = append(proxies, p)
	}
	return proxies
}

func (e *endpoint) DeleteProxy(id peer.ID) {
	e.Lock()
	defer e.Unlock()
	delete(e.proxies, id)
}

func (e *endpoint) DiscoveryProxies(ctx context.Context) ([]peer.ID, error) {
	addrs, err := discovery2.FindPeers(ctx, e.discoverer, e.cfg.ServiceTag)
	if err != nil {
		return nil, err
	}
	proxies := make([]peer.ID, 0, len(addrs))
	for _, addr := range addrs {
		proxies = append(proxies, addr.ID)
	}
	return proxies, nil
}

func (e *endpoint) Stop() error {
	close(e.stopping)
	errs := make([]error, 0, len(e.listeners)+1)
	for _, lsr := range e.listeners {
		errs = append(errs, lsr.Close())
	}
	errs = append(errs, e.node.Close())
	return multierr.Combine(errs...)
}
