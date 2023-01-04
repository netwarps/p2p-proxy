package p2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/diandianl/p2p-proxy/config"
	"github.com/diandianl/p2p-proxy/log"
)

func NewHost(ctx context.Context, cfg *config.Config) (h host.Host, err error) {
	logger := log.NewSubLogger("p2p")

	hostOpts, err := hostOptions(ctx, cfg)
	if err != nil {
		return
	}

	var bootPeers []ma.Multiaddr

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

	h, err = libp2p.New(hostOpts...)
	if err != nil {
		return
	}

	if len(h.Addrs()) > 0 {
		logger.Infof("P2P [/ipfs/%s] working addrs: %s", h.ID(), h.Addrs())
	}

	var wg sync.WaitGroup
	for _, peerAddr := range bootPeers {
		wg.Add(1)
		go func(addr ma.Multiaddr) {
			defer wg.Done()
			info, _ := peer.AddrInfoFromP2pAddr(addr)
			if err := h.Connect(ctx, *info); err != nil {
				logger.Warn(err)
			} else {
				logger.Debug("Connection established with bootstrap node:", info)
				_, err = permanentAddAddrToPeerstore(h, addr)
			}
		}(peerAddr)
	}
	wg.Wait()

	return
}

func convertToMAddr(addrs []string) ([]ma.Multiaddr, error) {
	maddrs := make([]ma.Multiaddr, 0, len(addrs))
	for _, addr := range addrs {
		ma, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		maddrs = append(maddrs, ma)
	}
	return maddrs, nil
}

func permanentAddAddrToPeerstore(h host.Host, addr ma.Multiaddr) (peer.ID, error) {
	pid, err := addr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return "", err
	}

	peerid, err := peer.Decode(pid)
	if err != nil {
		return "", err
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.Encode(peerid)))
	targetAddr := addr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr so we add
	// it to the peerstore so LibP2P knows how to contact it
	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid, nil
}
