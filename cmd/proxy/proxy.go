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
package proxy

import (
	"context"

	"github.com/diandianl/p2p-proxy/config"
	"github.com/diandianl/p2p-proxy/proxy"

	"github.com/spf13/cobra"

	_ "github.com/diandianl/p2p-proxy/protocol/service/http"
	_ "github.com/diandianl/p2p-proxy/protocol/service/socks5"
)

// proxyCmd represents the proxy command
func NewProxyCmd(ctx context.Context) *cobra.Command {
	var proxyCmd = &cobra.Command{
		Use:   "proxy",
		Short: "Start a proxy server peer",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			cfgFile := cmd.Flags().Lookup("config").Value.String()
			cfg, err := config.LoadOrInitializeIfNotPresent(cfgFile)
			if err != nil {
				return err
			}
			err = cfg.Validate(true)
			if err != nil {
				return err
			}
			err = cfg.SetLogLevel(cmd.Flags().Lookup("log-level").Value.String())
			if err != nil {
				return err
			}
			proxyService, err := proxy.New(cfg)
			if err != nil {
				return err
			}
			return proxyService.Start(ctx)
		},
	}
	return proxyCmd
}
