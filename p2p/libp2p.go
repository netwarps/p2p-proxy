package p2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/diandianl/p2p-proxy/config"
	"github.com/diandianl/p2p-proxy/log"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/kingwel-xie/xcli"
	"github.com/libp2p/go-libp2p"
	autonat "github.com/libp2p/go-libp2p-autonat-svc"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery2 "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dhtopts "github.com/libp2p/go-libp2p-kad-dht/opts"
	secio "github.com/libp2p/go-libp2p-secio"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	maddr "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"github.com/urfave/cli/v2"
)

func NewHostAndDiscovererAndBootstrap(ctx context.Context, cfg *config.Config) (h host.Host, dis discovery.Discovery, err error) {
	logger := log.NewSubLogger("p2p")

	hostOpts, err := hostOptions(ctx, cfg)
	if err != nil {
		return
	}

	dhtOpts, err := dhtOptions(cfg)
	if err != nil {
		return
	}

	var bootPeers []maddr.Multiaddr // dht.DefaultBootstrapPeers

	if len(cfg.P2P.BootstrapPeers) > 0 {
		bootPeers, err = convertToMAddr(cfg.P2P.BootstrapPeers)
		if err != nil {
			return
		}
	} else if len(config.Default.P2P.BootstrapPeers) > 0 {
		bootPeers, err = convertToMAddr(config.Default.P2P.BootstrapPeers)
		if err != nil {
			return
		}
	}

	dhtDs := dssync.MutexWrap(ds.NewMapDatastore())

	dhtOpts = append(dhtOpts, func(options *dhtopts.Options) error {
		options.Datastore = dhtDs
		return nil
	})

	if cfg.P2P.Identity.ObservedAddrActivationThresh > 0 {
		identify.ActivationThresh = cfg.P2P.Identity.ObservedAddrActivationThresh
		logger.Debugf("Override LibP2P observed address activation thresh with %d", identify.ActivationThresh)
	}

	dhtCreater := newDHTConstructor(ctx, dhtOpts)

	if cfg.P2P.EnableAutoRelay {
		logger.Debugf("Enable auto relay")
		hostOpts = append(hostOpts,
			libp2p.Routing(func(h host.Host) (routing routing.PeerRouting, err error) {
				return dhtCreater(h)
			}),
			libp2p.EnableAutoRelay(),
		)
	}

	h, err = libp2p.New(ctx, hostOpts...)
	if err != nil {
		return nil, nil, err
	}

	if len(h.Addrs()) > 0 {
		logger.Infof("P2P [/ipfs/%s] working addrs: %s", h.ID().Pretty(), h.Addrs())
	}

	if cfg.P2P.AutoNATService {
		logger.Debugf("Enable auto NAT service")
		_, err = autonat.NewAutoNATService(ctx, h,
			libp2p.Security(secio.ID, secio.New),
			libp2p.DefaultTransports,
		)
		if err != nil {
			h.Close()
			return nil, nil, err
		}
	}

	router, err := dhtCreater(h)
	if err != nil {
		h.Close()
		return nil, nil, err
	}

	err = router.Bootstrap(ctx)
	if err != nil {
		h.Close()
		return nil, nil, err
	}

	var connected []peer.AddrInfo

	var wg sync.WaitGroup
	for _, peerAddr := range bootPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()
			if err := h.Connect(ctx, pi); err != nil {
				logger.Warn(err)
			} else {
				connected = append(connected, pi)
				logger.Debug("Connection established with bootstrap node:", pi)
				_, err = permanentAddAddrToPeerstore(h, peerAddr)
				if err != nil {
					logger.Warnf("Permanent add addr [%s] to peerstore: ", peerAddr, err)
				}
			}
		}(*peerinfo)
	}
	wg.Wait()

	dis = discovery2.NewRoutingDiscovery(router)

	if cfg.Interactive {
		logger.Debug("Enable interactive mode")
		xcli.IPFS_PEERS = connected
		xcli.RunP2PNodeCLI(&xcli.P2PNode{
			Host: h,
			Dht:  router.(*dht.IpfsDHT),
			Ds:   dhtDs,
		},
			extCmds(),
		)
	}
	return h, dis, nil
}

func extCmds() cli.Commands {

	return cli.Commands{
		{
			Name:        "a2cid",
			Category:    "p2p",
			Usage:       "a2cid <str>",
			Description: "calc str cid",
			Action: func(c *cli.Context) error {
				cid, err := a2cid(c.Args().First())
				if err != nil {
					return err
				}
				fmt.Fprintln(c.App.Writer, cid.Hash().B58String())
				return nil
			},
		},
	}
}

func a2cid(str string) (cid.Cid, error) {
	h, err := mh.Sum([]byte(str), mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}
	return cid.NewCidV1(cid.Raw, h), nil
}

func permanentAddAddrToPeerstore(h host.Host, ma maddr.Multiaddr) (peer.ID, error) {
	pid, err := ma.ValueForProtocol(maddr.P_IPFS)
	if err != nil {
		return "", err
	}

	peerid, err := peer.Decode(pid)
	if err != nil {
		return "", err
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := maddr.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.Encode(peerid)))
	targetAddr := ma.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr so we add
	// it to the peerstore so LibP2P knows how to contact it
	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid, nil
}

func newDHTConstructor(ctx context.Context, opts []dhtopts.Option) func(host.Host) (routing.Routing, error) {
	var (
		once   sync.Once
		router routing.Routing
		err    error
	)
	return func(h host.Host) (routing.Routing, error) {
		once.Do(func() {
			router, err = dht.New(ctx, h, opts...)
		})
		return router, err
	}
}

func convertToMAddr(addrs []string) ([]maddr.Multiaddr, error) {
	maddrs := make([]maddr.Multiaddr, 0, len(addrs))
	for _, addr := range addrs {
		ma, err := maddr.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		maddrs = append(maddrs, ma)
	}
	return maddrs, nil
}
