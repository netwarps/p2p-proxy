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
	"crypto/rand"
	"encoding/base64"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "generate and write default config",
	Long: "generate and write default config",
	RunE: func(cmd *cobra.Command, args []string) error {
		// parent command PersistentPreRunE hook ensure configFile exist
		return initConfig("")
	},
}

func initConfig(saveAs string) error {

	log.Debug("Initial run or specified configuration file does not exist or reinit, perform initialization")

	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return err
	}
	privKey, err := crypto.MarshalPrivateKey(priv)

	if err != nil {
		return err
	}

	log.Debug("Generated Private Key, Using Alg: RSA, Bits: 2048")

	viper.Set("Identity.PrivKey", base64.StdEncoding.EncodeToString(privKey))

	addrs := map[string]interface{}{
		"Endpoint.HTTP": "127.0.0.1:8010",
		"Endpoint.Proxy": "/ip4/149.129.82.89/tcp/8888/ipfs/QmXwj9Uk68XTGZLQrREjQJpTLx6GWokHrGX7xrYPGcRkTn",
		"P2P.Addr": []string{"/ip4/0.0.0.0/tcp/8888"},
	}

	for key, def := range addrs {
		if !viper.IsSet(key) {
			viper.Set(key, def)
		}
	}
	if saveAs != "" {
		// config file not exist
		return viper.WriteConfigAs(saveAs)
	}
	return viper.WriteConfig()
}
