package roundrobin

import (
	"github.com/diandianl/p2p-proxy/endpoint/balancer"
	"github.com/diandianl/p2p-proxy/protocol"

	"go.uber.org/atomic"
)

func init() {
	err := balancer.RegisterBalancerFactory(balancer.RoundRobin, New)
	if err != nil {
		panic(err)
	}
}

func New(getter balancer.Getter) (balancer.Balancer, error) {
	return &roundRobin{Getter: getter}, nil
}

type roundRobin struct {
	balancer.Getter
	counter atomic.Uint32
}

func (rr *roundRobin) Name() string {
	return balancer.RoundRobin
}

func (rr *roundRobin) Next(p protocol.Protocol) (balancer.Proxy, error) {
	proxies := rr.GetProxies(p)
	if len(proxies) == 0 {
		return balancer.NoProxy, balancer.NewNotEnoughProxiesError(p)
	}
	if len(proxies) == 1 {
		return proxies[0], nil
	}
	return proxies[rr.counter.Inc()%uint32(len(proxies))], nil
}
