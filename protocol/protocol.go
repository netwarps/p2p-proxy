package protocol

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/diandianl/p2p-proxy/log"
)

const (
	HTTP Protocol = "/p2p-proxy/http/0.1.0"

	Shadowsocks Protocol = "/p2p-proxy/shadowsocks/0.1.0"
)

type Protocol string

type Service interface {
	Protocol() Protocol

	Serve(context.Context, net.Listener) error

	Shutdown(context.Context) error
}

type ServiceFactory func(logger log.Logger, cfg map[string]interface{}) (Service, error)

type Listener interface {
	io.Closer

	Protocol() Protocol

	Accept() (net.Conn, error)
}

type ListenerFactory func(logger log.Logger, listen string) (Listener, error)

type metadata struct {
	protocol   Protocol
	short      string
	svcFactory ServiceFactory
	lsrFactory ListenerFactory
}

var svcRegistry = map[Protocol]*metadata{}

func RegisterServiceFactory(protocol Protocol, short string, factory ServiceFactory) error {
	if _, ok := svcRegistry[protocol]; ok {
		return fmt.Errorf("duplicate registration, Protocol [%s] Factory registered", protocol)
	}
	svcRegistry[protocol] = &metadata{protocol: protocol, short: short, svcFactory: factory}
	return nil
}

func NewService(protocol Protocol, cfg map[string]interface{}) (Service, error) {
	m, ok := svcRegistry[protocol]
	if !ok {
		if len(svcRegistry) == 0 {
			return nil, fmt.Errorf("you should first call 'func RegisterServiceFactory' to register")
		}
		return nil, fmt.Errorf("unsupported Protocol [%s]", protocol)
	}
	logger := log.NewSubLogger(m.short)
	s, err := m.svcFactory(logger, cfg)
	if err != nil {
		return nil, err
	}
	if protocol != s.Protocol() {
		return nil, fmt.Errorf("mismatched protocol, expect [%s], got [%s]", protocol, s.Protocol())
	}
	return s, nil
}

var lsrRegistry = map[Protocol]*metadata{}

func RegisterListenerFactory(protocol Protocol, short string, factory ListenerFactory) error {
	if _, ok := lsrRegistry[protocol]; ok {
		return fmt.Errorf("duplicate registration, Protocol [%s] Factory registered", protocol)
	}
	lsrRegistry[protocol] = &metadata{protocol: protocol, short: short, lsrFactory: factory}
	return nil
}

func NewListener(protocol Protocol, listen string) (Listener, error) {
	m, ok := lsrRegistry[protocol]
	if !ok {
		if len(svcRegistry) == 0 {
			return nil, fmt.Errorf("you should first call 'func RegisterListenerFactory' to register")
		}
		return nil, fmt.Errorf("unsupported Protocol [%s]", protocol)
	}
	logger := log.NewSubLogger(m.short)
	s, err := m.lsrFactory(logger, listen)
	if err != nil {
		return nil, err
	}
	if protocol != s.Protocol() {
		return nil, fmt.Errorf("mismatched protocol, expect [%s], got [%s]", protocol, s.Protocol())
	}
	return s, nil
}
