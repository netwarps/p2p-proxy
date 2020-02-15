package p2p

import (
	"context"
	"fmt"
	"time"

	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/metrics"
)

var log = logging.Logger("p2p-proxy")

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

func BandwidthReporter(ctx context.Context, period time.Duration) (libp2p.Option, error) {
	reporter := metrics.NewBandwidthCounter()
	ticker := time.NewTicker(period)

	go func() {
		for {
			select {
			case <-ticker.C:
				stats := reporter.GetBandwidthTotals()
				log.Infof("BW Speed: TIN %s, TOUT %s, RIN %s, ROUT %s\n",
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

	return libp2p.BandwidthReporter(reporter), nil
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
