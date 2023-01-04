package balancer

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/diandianl/p2p-proxy/protocol"
)

const (
	NoProxy Proxy = ""

	RoundRobin = "round_robin"
)

type Proxy = peer.ID

type Getter interface {
	GetProxies(protocol protocol.Protocol) []Proxy
}

type Balancer interface {
	Name() string
	Next(protocol protocol.Protocol) (Proxy, error)
}

type BalancerFactory func(getter Getter) (Balancer, error)

var registry = map[string]BalancerFactory{}

func RegisterBalancerFactory(name string, factory BalancerFactory) error {
	if _, ok := registry[name]; ok {
		return fmt.Errorf("duplicate registration, name [%s] balancer factory registered", name)
	}
	registry[name] = factory
	return nil
}

func New(name string, getter Getter) (Balancer, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unsupported balancer [%s]", name)
	}
	return factory(getter)
}

type notEnoughProxiesError struct {
	p protocol.Protocol
}

func (e notEnoughProxiesError) Error() string {
	return fmt.Sprintf("not enough proxies for protocol [%s]", e.p)
}
func NewNotEnoughProxiesError(p protocol.Protocol) error {
	return notEnoughProxiesError{p}
}
func IsNewNotEnoughProxiesError(e error) bool {
	_, ok := e.(notEnoughProxiesError)
	return ok
}
