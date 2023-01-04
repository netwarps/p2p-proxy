package http

import (
	"context"
	"net"
	"net/http"

	"github.com/diandianl/p2p-proxy/log"
	"github.com/diandianl/p2p-proxy/protocol"

	"github.com/elazarl/goproxy"
)

func init() {
	err := protocol.RegisterServiceFactory(protocol.HTTP, "http", New)
	if err != nil {
		panic(err)
	}
}

func New(logger log.Logger, cfg map[string]interface{}) (protocol.Service, error) {
	proxy := goproxy.NewProxyHttpServer()

	setLogger(proxy, logger)

	return &goproxyService{logger: logger, srv: &http.Server{Handler: proxy}, delegate: proxy}, nil
}

type goproxyService struct {
	logger log.Logger

	srv *http.Server

	delegate *goproxy.ProxyHttpServer
}

func (_ *goproxyService) Protocol() protocol.Protocol {
	return protocol.HTTP
}

func (s *goproxyService) Serve(ctx context.Context, l net.Listener) error {
	err := s.srv.Serve(l)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *goproxyService) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
