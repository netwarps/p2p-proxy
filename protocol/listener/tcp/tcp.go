package tcp

import (
	"github.com/diandianl/p2p-proxy/log"
	"github.com/diandianl/p2p-proxy/protocol"
	"net"
)

func init() {
	err := protocol.RegisterListenerFactory(NewFactory(protocol.HTTPProtocol, "http"))
	if err != nil {
		panic(err)
	}
	err = protocol.RegisterListenerFactory(NewFactory(protocol.Socks5Protocol, "socks5"))
	if err != nil {
		panic(err)
	}
}

func NewFactory(p protocol.Protocol, short string) (protocol.Protocol, string, func(logger log.Logger, listen string) (protocol.Listener, error)) {
	return p, short, func(logger log.Logger, listen string) (protocol.Listener, error) {
		l, err := net.Listen("tcp", listen)
		if err != nil {
			return nil, err
		}
		return &listener{protocol: p, Listener: l}, nil
	}
}

type listener struct {
	protocol protocol.Protocol

	net.Listener
}

func (l *listener) Protocol() protocol.Protocol {
	return l.protocol
}
