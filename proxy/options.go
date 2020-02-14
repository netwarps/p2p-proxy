package proxy

import (
	"time"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
	"github.com/libp2p/go-libp2p"
)

type config struct {

	dialTimeout time.Duration

	// libp2p options
	libp2pOptions []libp2p.Option

	// goproxy options
	goproxyOptions []GoProxyOption
}

type Option func(*config) error

func AddP2POption(opt libp2p.Option, err error) Option {
	return Option(func(c *config) error{
		if err != nil {
			return err
		}
		c.libp2pOptions = append(c.libp2pOptions, opt)
		return nil
	})
}

func AddGoProxyOptions(opts ...GoProxyOption) Option {
	return Option(func(c *config) error {
		c.goproxyOptions = append(c.goproxyOptions, opts...)
		return nil
	})
}

type GoProxyOption func(server *goproxy.ProxyHttpServer) error

func BasicAuth(realm string, passwds map[string]string) GoProxyOption {
	return GoProxyOption(func(server *goproxy.ProxyHttpServer) error {
		auth.ProxyBasic(server, realm, func(user, passwd string) bool {
			if p, ok := passwds[user]; ok && p == passwd {
				return true
			}
			return false
		})
		return nil
	})
}

func LoggerAdapter() GoProxyOption {
	return func(server *goproxy.ProxyHttpServer) error {
		server.Logger = NewLoggerAdapter(server)
		return nil
	}
}


func (cfg *config) Libp2pOptions() []libp2p.Option {
	return cfg.libp2pOptions
}

func (cfg *config) applyGoProxyOptions(server *goproxy.ProxyHttpServer) error {
	for _, o := range cfg.goproxyOptions {
		if err := o(server); err != nil {
			return err
		}
	}
	return nil
}
