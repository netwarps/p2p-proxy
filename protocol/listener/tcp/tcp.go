package tcp

import (
	"github.com/diandianl/p2p-proxy/log"
	"github.com/diandianl/p2p-proxy/protocol"
	"net"
)

func init() {
	protos := []struct {
		protocol protocol.Protocol
		short    string
	}{
		{protocol.HTTP, "http"},
		{protocol.Shadowsocks, "shadowsocks"},
	}
	for _, proto := range protos {
		err := protocol.RegisterListenerFactory(NewFactory(proto.protocol, proto.short))
		if err != nil {
			panic(err)
		}
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
