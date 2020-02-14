package proxy

import (
	"context"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"go.uber.org/multierr"
)

// Protocol defines the libp2p protocol that we will use for the libp2p proxy
// service that we are going to provide. This will tag the streams used for
// this service. Streams are multiplexed and their protocol tag helps
// libp2p handle them to the right handler functions.
const Protocol = "/proxy-example/0.0.1"

var log = logging.Logger("p2p-proxy")

type ProxyService interface {

	Start(context.Context) error

	Stop() error

}

func New(opts ...Option) (ProxyService, error) {
	cfg := &config{}
	for _, o := range opts {
		if err := o(cfg); err != nil {
			return nil, err
		}
	}
	return &proxyService{ cfg: cfg, }, nil
}

type proxyService struct {
	cfg *config

	host host.Host

	target *goproxy.ProxyHttpServer

	srv *http.Server
}

func (ps *proxyService) Start(ctx context.Context) (err error) {

	log.Debug("Starting Proxy Server")

	opts := ps.cfg.Libp2pOptions()

	ps.host, err = libp2p.New(ctx, opts...)
	if err != nil {
		return err
	}

	l, err := gostream.Listen(ps.host, Protocol)
	if err != nil {
		return err
	}

	proxy := goproxy.NewProxyHttpServer()

	err = ps.cfg.applyGoProxyOptions(proxy)

	if err != nil {
		return err
	}

	ps.srv = &http.Server{Handler: proxy}

	log.Info("Proxy server is ready")
	log.Info("libp2p-peer addresses:")
	for _, a := range ps.host.Addrs() {
		log.Infof("\t%s/ipfs/%s\n", a, peer.Encode(ps.host.ID()))
	}

	go func() {
		<- ctx.Done()
		err := ps.Stop()
		if err != nil {
			log.Error("Stop Proxy Server: ", err)
		}
	}()

	err = ps.srv.Serve(l)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (ps *proxyService) Stop() error {
	//使用context控制srv.Shutdown的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return multierr.Combine(ps.srv.Shutdown(ctx), ps.host.Close())
}
