package socks5

import (
	"context"
	"github.com/diandianl/p2p-proxy/log"
	"github.com/diandianl/p2p-proxy/protocol"
	"net"

	socks5 "github.com/armon/go-socks5"
)

func init() {
	err := protocol.RegisterServiceFactory(protocol.Socks5, "socks5", New)
	if err != nil {
		panic(err)
	}
}

func New(logger log.Logger, cfg map[string]interface{}) (protocol.Service, error) {

	// TODO process cfg
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		return nil, err
	}
	return &socks5Service{logger: logger, delegate: server}, nil
}

type socks5Service struct {
	logger log.Logger

	delegate *socks5.Server

	listener net.Listener

	shuttingDown bool
}

func (_ *socks5Service) Protocol() protocol.Protocol {
	return protocol.Socks5
}

func (s *socks5Service) Serve(ctx context.Context, l net.Listener) error {
	s.listener = l
	err := s.delegate.Serve(l)
	if s.shuttingDown {
		err = nil
	}
	return err
}

func (s *socks5Service) Shutdown(ctx context.Context) error {
	s.shuttingDown = true
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
