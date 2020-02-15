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
	"path"

	logging "github.com/ipfs/go-log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var log = logging.Logger("p2p-proxy")

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {

	cmd := newCommand(ctx)

	if err := cmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

// new root command
func newCommand(ctx context.Context) *cobra.Command {
	var cfgFile string
	var logLevel string

	ep := newEndpointCmd(ctx)

	cmd := &cobra.Command{
		Use:   "p2p-proxy",
		Short: "A http(s) proxy based on P2P",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			lvl, err := logging.LevelFromString(logLevel)
			if err != nil {
				return err
			}
			logging.SetAllLoggers(lvl)

			cfgFile, err := loadConfig(cfgFile)
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				return initConfig(cfgFile)
			}
			return err
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

	cmd.AddCommand(initCmd)

	cmd.AddCommand(newProxyCmd(ctx))

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func loadConfig(cfgFile string) (string, error) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			return cfgFile, err
		}
		// Search config in home directory with name ".p2p-proxy" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".p2p-proxy")
		viper.SetConfigType("yaml")

		cfgFile = path.Join(home, ".p2p-proxy.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	err := viper.ReadInConfig()
	if err != nil {
		return cfgFile, err
	}
	// If a config file is found, read it in.
	log.Info("Using config file:", viper.ConfigFileUsed())
	return cfgFile, nil
}
