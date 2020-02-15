/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/diandianl/p2p-proxy/p2p"
	"github.com/diandianl/p2p-proxy/proxy"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// proxyCmd represents the proxy command
func newProxyCmd(ctx context.Context) *cobra.Command {
	var proxyCmd = &cobra.Command{
		Use:   "proxy",
		Short: "Start a proxy server peer",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			checkCfgs := []string{
				"Identity.PrivKey",
			}

			for _, k := range checkCfgs {
				if !viper.IsSet(k) {
					return fmt.Errorf("Config '%s' is required", k)
				}
			}

			viper.SetDefault("P2P.Addr", []string{"/ip4/127.0.0.1/tcp/8888"})

			priv := viper.GetString("Identity.PrivKey")
			privKey, err := base64.StdEncoding.DecodeString(priv)
			if err != nil {
				return err
			}

			opts := []proxy.Option{
				proxy.AddP2POption(p2p.Identity(privKey)),
				proxy.AddP2POption(p2p.Addrs(viper.GetStringSlice("P2P.Addr")...)),
			}

			var goproxyOptions = []proxy.GoProxyOption{proxy.LoggerAdapter() }

			if viper.IsSet("Proxy.Auth.Basic.Realm") && viper.IsSet("Proxy.Auth.Basic.Users") {
				realm := viper.GetString("Proxy.Auth.Basic.Realm")
				users := viper.GetStringMapString("Proxy.Auth.Basic.Users")
				log.Debugf("Enable basic auth, with realm '%s'", realm)
				goproxyOptions = append(goproxyOptions, proxy.BasicAuth(realm, users))
			}

			opts = append(opts, proxy.AddGoProxyOptions(goproxyOptions...))

			if viper.GetBool("P2P.BandwidthReporter.Enable") {
				period := 10 * time.Second
				if viper.IsSet("P2P.BandwidthReporter.Period") {
					period = viper.GetDuration("P2P.BandwidthReporter.Period")
				}
				opts = append(opts, proxy.AddP2POption(p2p.BandwidthReporter(ctx, period)))
			}

			proxyService, err := proxy.New(opts...)
			if err != nil {
				return err
			}
			return proxyService.Start(ctx)
		},
	}

	return proxyCmd
}
