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
	"os"

	"github.com/diandianl/p2p-proxy/cmd/endpoint"
	"github.com/diandianl/p2p-proxy/cmd/proxy"
	"github.com/diandianl/p2p-proxy/config"
	"github.com/diandianl/p2p-proxy/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {

	viper.SetDefault("Version", config.Version)

	logger := log.NewLogger()

	cmd := newCommand(ctx, logger)

	if err := cmd.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}

// new root command
func newCommand(ctx context.Context, logger log.Logger) *cobra.Command {
	var cfgFile string
	var logLevel string

	ep := endpoint.NewEndpointCmd(ctx)

	cmd := &cobra.Command{
		Use:   "p2p-proxy",
		Short: "A http(s) proxy based on P2P",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return log.SetAllLogLevel(logLevel)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return ep.RunE(cmd, args)
		},
	}

	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "INFO", "set logging level")

	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.p2p-proxy.yaml)")

	cmd.PersistentFlags().StringSlice("p2p-addr", []string{}, "peer listen addr(s)")
	viper.BindPFlag("P2P.Addr", cmd.PersistentFlags().Lookup("p2p-addr"))

	cmd.Flags().AddFlagSet(ep.Flags())

	// cmd.AddCommand(initCmd)

	cmd.AddCommand(proxy.NewProxyCmd(ctx))

	return cmd
}
