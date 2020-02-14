package p2p

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
)

func Identity(privKey []byte) (libp2p.Option, error) {
	priv, err := crypto.UnmarshalPrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	return libp2p.Identity(priv), nil
}

func Addrs(addrs ...string) (libp2p.Option, error) {
	return libp2p.ListenAddrStrings(addrs...), nil
}
