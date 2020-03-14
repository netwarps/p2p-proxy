package p2p

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/diandianl/p2p-proxy/config"
	"github.com/diandianl/p2p-proxy/log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/metrics"
	dhtopts "github.com/libp2p/go-libp2p-kad-dht/opts"
)

func hostOptions(ctx context.Context, c *config.Config) (opts []libp2p.Option, err error) {
	var opt libp2p.Option
	if opt, err = identity(c.P2P.Identity.PrivKey); err != nil {
		return
	}
	opts = append(opts, opt)

	if !(c.Work4Proxy() || c.P2P.AutoNATService || c.P2P.EnableAutoRelay) {
		opts = append(opts, libp2p.NoListenAddrs)
	} else {
		if opt, err = listenAddrs(c.P2P.Addrs...); err != nil {
			return
		}
		opts = append(opts, opt)
	}

	if c.P2P.BandWidthReporter.Enable {
		if opt, err = BandwidthReporter(ctx, c.P2P.BandWidthReporter.Interval); err != nil {
			return
		}
		opts = append(opts, opt)
	}
	return
}

func dhtOptions(c *config.Config) ([]dhtopts.Option, error) {
	return []dhtopts.Option{dhtopts.Client(c.P2P.DHT.Client)}, nil
}

func identity(privKey string) (libp2p.Option, error) {
	priv, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return nil, err
	}
	pk, err := crypto.UnmarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	return libp2p.Identity(pk), nil
}

func listenAddrs(addrs ...string) (libp2p.Option, error) {
	return libp2p.ListenAddrStrings(addrs...), nil
}

func BandwidthReporter(ctx context.Context, period time.Duration) (libp2p.Option, error) {
	logger := log.NewSubLogger("reporter")

	counter := metrics.NewBandwidthCounter()
	ticker := time.NewTicker(period)

	go func() {
		for {
			select {
			case <-ticker.C:
				stats := counter.GetBandwidthTotals()
				logger.Infof("BW Speed: TIN %s, TOUT %s, RIN %s, ROUT %s\n",
					byteCountBinary(stats.TotalIn),
					byteCountBinary(stats.TotalOut),
					byteRateBinary(stats.RateIn),
					byteRateBinary(stats.RateOut))
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	return libp2p.BandwidthReporter(counter), nil
}

/*
func byteCountDecimal(b int64) string {
        const unit = 1000
        if b < unit {
                return fmt.Sprintf("%d B", b)
        }
        div, exp := int64(unit), 0
        for n := b / unit; n >= unit; n /= unit {
                div *= unit
                exp++
        }
        return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
*/

func byteCountBinary(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func byteRateBinary(b float64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%.1f B/s", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB/s", b/float64(div), "KMGTPE"[exp])
}
