package endpoint

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/diandianl/p2p-proxy/proxy"

	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/multierr"
)

var log = logging.Logger("p2p-proxy")

type Endpoint interface {

	Start(ctx context.Context) error

	Stop() error
}

func New(opts ...Option) (Endpoint, error) {

	cfg := &config{}

	for _, o := range opts {
		if err := o(cfg); err != nil {
			return nil, err
		}
	}

	return &endpoint{ cfg: cfg, stopping: make(chan struct{}) }, nil
}

type endpoint struct {
	cfg *config
	// remote proxy server peer
	dest peer.ID
	// p2p host
	host host.Host
	// http(s) proxy endpoint listener
	listener net.Listener

	stopping chan struct{}
}

func (e *endpoint) Start(ctx context.Context) (err error) {

	log.Debug("Starting Endpoint")

	opts := e.cfg.Libp2pOptions()

	e.host, err = libp2p.New(ctx, opts...)
	if err != nil {
		return err
	}

	e.dest, err = addAddrToPeerstore(e.host, e.cfg.proxy)
	if err != nil {
		return err
	}

	log.Info("proxy listening on ", e.cfg.listen)
	e.listener, err = net.Listen("tcp", e.cfg.listen)
	if err != nil {
		return err
	}

	go func() {
		<- ctx.Done()
		err := e.Stop()
		if err != nil {
			log.Error("Stop Endpoint: ", err)
		}
	}()

	for {
		conn, err := e.listener.Accept()
		if err != nil {
			select{
			case <- e.stopping:
				return nil
			default:
			}
			return err
		}
		go func(){
			e.connHandler(conn)
		}()
	}
}

func (e *endpoint) connHandler(conn net.Conn) {

	log.Debug("New Conn", conn.RemoteAddr())

	defer conn.Close()
	stream, err := e.host.NewStream(context.Background(), e.dest, proxy.Protocol)
	// If an error happens, we write an error for response.
	if err != nil {
		log.Error("New Stream", err)
		return
	}
	defer stream.Close()

	var wg sync.WaitGroup

	wg.Add(2)

	go func () {
		defer wg.Done()
		_, err := io.Copy(stream, conn)
		if err != nil {
			conn.Close()
			stream.Reset()
			log.Error("Copy stream => conn: ", err)
			return
		}
	}()
	go func () {
		defer wg.Done()
		_, err := io.Copy(conn, stream)
		if err != nil {
			conn.Close()
			stream.Reset()
			log.Error("Copy conn => stream: ", err)
			return
		}
	}()

	wg.Wait()
}

func (e *endpoint) Stop() error {
	close(e.stopping)
	return multierr.Combine(e.listener.Close(), e.host.Close())
}

// addAddrToPeerstore parses a peer multiaddress and adds
// it to the given host's peerstore, so it knows how to
// contact it. It returns the peer ID of the remote peer.
func addAddrToPeerstore(h host.Host, addr string) (peer.ID, error) {
	// The following code extracts target's the peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return "", err
	}
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
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
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr so we add
	// it to the peerstore so LibP2P knows how to contact it
	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid, nil
}
