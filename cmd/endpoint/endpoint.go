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
package endpoint

import (
	"context"

	"github.com/diandianl/p2p-proxy/config"
	"github.com/diandianl/p2p-proxy/endpoint"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/diandianl/p2p-proxy/endpoint/balancer/roundrobin"
	_ "github.com/diandianl/p2p-proxy/protocol/listener/tcp"
)

func NewEndpointCmd(ctx context.Context, cfgGetter func(proxy bool) (*config.Config, error)) *cobra.Command {

	// endpointCmd represents the endpoint command
	var endpointCmd = &cobra.Command{
		Use:   "endpoint",
		Short: "endpoint command run at local for proxy agent",
		Long:  "endpoint command run at local for proxy agent",
		RunE: func(cmd *cobra.Command, args []string) error {

			cmd.SilenceUsage = true
			cfg, err := cfgGetter(false)
			if err != nil {
				return err
			}

			ep, err := endpoint.New(cfg)

			if err != nil {
				return err
			}

			return ep.Start(ctx)
		},
	}

	endpointCmd.Flags().StringP("proxy", "p", "", "proxy server address")
	viper.BindPFlag("Endpoint.Proxy", endpointCmd.Flags().Lookup("proxy"))

	endpointCmd.Flags().String("http", "", "local http(s) proxy agent listen address")
	viper.BindPFlag("Endpoint.HTTP", endpointCmd.Flags().Lookup("http"))

	return endpointCmd
}
