package endpoint

import "github.com/libp2p/go-libp2p"

type config struct {
	listen string
	proxy string

	// libp2p options
	libp2pOptions []libp2p.Option
}

type Option func(*config) error

func Listen(addr string) Option {
	return Option(func(c *config) error {
		c.listen = addr
		return nil
	})
}

func Proxy(addr string) Option {
	return Option(func(c *config) error {
		c.proxy = addr
		return nil
	})
}

func AddP2POption(opt libp2p.Option, err error) Option {
	return Option(func(c *config) error{
		if err != nil {
			return err
		}
		c.libp2pOptions = append(c.libp2pOptions, opt)
		return nil
	})
}

func (cfg *config) Libp2pOptions() []libp2p.Option {
	return cfg.libp2pOptions
}

