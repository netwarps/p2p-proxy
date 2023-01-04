package proxy

import (
	"context"
	"errors"

	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/host"
	p2pproto "github.com/libp2p/go-libp2p/core/protocol"
	"go.uber.org/multierr"

	cfg "github.com/diandianl/p2p-proxy/config"
	"github.com/diandianl/p2p-proxy/log"
	"github.com/diandianl/p2p-proxy/p2p"
	"github.com/diandianl/p2p-proxy/protocol"
)

type Server struct {
	logger log.Logger

	cfg *cfg.Config

	node host.Host

	services []protocol.Service
}

func New(cfg *cfg.Config) (*Server, error) {
	if err := cfg.Validate(true); err != nil {
		return nil, err
	}
	return &Server{logger: log.NewSubLogger("proxy"), cfg: cfg}, nil
}

func (s *Server) Start(ctx context.Context) error {

	logger := s.logger
	defer logger.Sync()

	logger.Infof("Starting Proxy Server")

	c := s.cfg

	if len(c.Proxy.Protocols) == 0 {
		return errors.New("'Config.Proxy.Protocols' can not be empty")
	}

	h, err := p2p.NewHost(ctx, c)
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

	<-ctx.Done()
	return s.Stop()
}

func (s *Server) startService(ctx context.Context, svc protocol.Service) error {
	l, err := gostream.Listen(s.node, p2pproto.ID(svc.Protocol()))
	if err != nil {
		return err
	}
	return svc.Serve(ctx, l)
}

func (s *Server) Stop() error {
	ctx := context.Background()
	errs := make([]error, 0, len(s.services)+1)
	for _, svc := range s.services {
		errs = append(errs, svc.Shutdown(ctx))
	}
	errs = append(errs, s.node.Close())
	return multierr.Combine(errs...)
}
