package proxy

import (
	"context"
	"errors"

	cfg "github.com/diandianl/p2p-proxy/config"
	"github.com/diandianl/p2p-proxy/log"
	"github.com/diandianl/p2p-proxy/p2p"
	"github.com/diandianl/p2p-proxy/protocol"

	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	p2pproto "github.com/libp2p/go-libp2p-core/protocol"
	discovery2 "github.com/libp2p/go-libp2p-discovery"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"go.uber.org/multierr"
)

type ProxyServer interface {
	Start(ctx context.Context) error

	Stop() error
}

func New(cfg *cfg.Config) (ProxyServer, error) {
	if err := cfg.Validate(true); err != nil {
		return nil, err
	}
	return &proxyServer{logger: log.NewSubLogger("proxy"), cfg: cfg}, nil
}

type proxyServer struct {
	logger log.Logger

	cfg *cfg.Config

	node host.Host

	services []protocol.Service
}

func (s *proxyServer) Start(ctx context.Context) error {

	logger := s.logger
	defer logger.Sync()

	logger.Infof("Starting Proxy Server")

	c := s.cfg

	if len(c.Proxy.Protocols) == 0 {
		return errors.New("'Config.Proxy.Protocols' can not be empty")
	}

	h, rd, err := p2p.NewHostAndDiscovererAndBootstrap(ctx, c)
	if err != nil {
		return err
	}
	s.node = h

	for _, proto := range c.Proxy.Protocols {
		svc, err := protocol.NewService(protocol.Protocol(proto.Protocol), proto.Config)
		if err != nil {
			return err
		}
		s.services = append(s.services, svc)
		logger.Infof("Supporting %s service", svc.Protocol())
		go func() {
			err := s.startService(ctx, svc)
			if err != nil {
				s.logger.Errorf("start proxy service [%s], ", svc.Protocol(), err)
			}
		}()
	}

	discovery2.Advertise(ctx, rd, c.ServiceTag, discovery.TTL(c.Proxy.ServiceAdvertiseInterval))

	<-ctx.Done()
	return s.Stop()
}

func (s *proxyServer) startService(ctx context.Context, svc protocol.Service) error {
	l, err := gostream.Listen(s.node, p2pproto.ID(svc.Protocol()))
	if err != nil {
		return err
	}
	return svc.Serve(ctx, l)
}

func (s *proxyServer) Stop() error {
	ctx := context.Background()
	errs := make([]error, 0, len(s.services)+1)
	for _, svc := range s.services {
		errs = append(errs, svc.Shutdown(ctx))
	}
	errs = append(errs, s.node.Close())
	return multierr.Combine(errs...)
}
