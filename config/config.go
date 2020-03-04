package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/diandianl/p2p-proxy/log"
	"github.com/libp2p/go-libp2p-core/crypto"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	DefaultConfigPath = "~/.p2p-proxy.yaml"
)

var Version = "v0.0.2"

var InvalidErr = errors.New("config invalid or not checked")

var Default = &Config{
	P2P: P2P{
		Addrs: []string{
			"/ip4/0.0.0.0/tcp/8888",
		},
		BootstrapPeers: []string{
		},
	},
	LogLevel: map[string]string{
		"all": "info",
	},
	Version:    Version,
	ServiceTag: "p2p-proxy/0.0.1",
	Proxy: Proxy{
		Protocols: []ProxyProtocol{
			{
				Protocol: "/p2p-proxy/http/0.0.1",
				Config:   map[string]interface{}{},
			},
			/*
				{
					Protocol: "/p2p-proxy/socks5/0.0.1",
					Config: map[string]interface{}{},
				},
			*/
		},
		ServiceAdvertiseInterval: time.Hour,
	},
	Endpoint: Endpoint{
		ProxyProtocols: []ProxyProtocol{
			{
				Protocol: "/p2p-proxy/http/0.0.1",
				Listen:   "127.0.0.1:8010",
			},
			/*
				{
					Protocol: "/p2p-proxy/socks5/0.0.1",
					Listen: "127.0.0.1:8020",
				},
			*/
		},

		ServiceDiscoveryInterval: time.Hour,

		Balancer: "round_robin",
	},
	Interactive: false,
}

func LoadOrInitializeIfNotPresent(configPath string) (*Config, error) {
	if len(configPath) == 0 {
		configPath = DefaultConfigPath
	}

	configPath, err := homedir.Expand(filepath.Clean(configPath))
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Initialize(configPath)
		}
		return nil, err
	}
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	if viper.GetString("Version") == "v0.0.1" {
		cfg.P2P.Identity.PrivKey = viper.GetString("Identity.PrivKey")
		cfg.P2P.Addrs = viper.GetStringSlice("P2P.Addr")

		cfg.Endpoint = Default.Endpoint
		cfg.Proxy = Default.Proxy
		cfg.ServiceTag = Default.ServiceTag
		cfg.Version = Version
		return writeConfig(configPath, cfg)
	}
	return cfg, nil
}

func Initialize(cfgPath string) (*Config, error) {
	var cfg Config = *Default

	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return nil, err
	}
	privKey, err := crypto.MarshalPrivateKey(priv)

	if err != nil {
		return nil, err
	}
	cfg.P2P.Identity.PrivKey = base64.StdEncoding.EncodeToString(privKey)

	fmt.Println(len(Default.P2P.Identity.PrivKey))

	return writeConfig(cfgPath, &cfg)
}

func writeConfig(configPath string, cfg *Config) (*Config, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(configPath, data, 0755)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

type Config struct {
	// config version
	Version string `yaml:"Version"`

	LogLevel map[string]string `yaml:"LogLevel"`

	ServiceTag string `yaml:"ServiceTag"`

	P2P P2P `yaml:"P2P"`

	Proxy Proxy `yaml:"Proxy"`

	Endpoint Endpoint `yaml:"Endpoint"`

	Interactive bool `yaml:"Interactive"`

	valid bool `yaml:"-"`
}

func (c *Config) Validate(proxy bool) error {
	if len(c.P2P.Identity.PrivKey) == 0 {
		return fmt.Errorf("no 'P2P.Identity.PrivKey' config")
	}

	if len(c.P2P.Addrs) == 0 {
		return fmt.Errorf("no 'P2P.Addrs' config")
	}
	if proxy {
		if len(c.Proxy.Protocols) == 0 {
			return fmt.Errorf("no 'Proxy.Protocols' config")
		}
	} else {
		if len(c.Endpoint.ProxyProtocols) == 0 {
			return fmt.Errorf("no 'Endpoint.ProxyProtocols' config")
		}
		if len(c.Endpoint.Balancer) == 0 {
			return fmt.Errorf("no 'Endpoint.Balancer' config")
		}
	}
	c.valid = true
	return nil
}

func (c *Config) SetLogLevel(defaultLevel string) error {
	if !c.valid {
		return InvalidErr
	}
	if len(c.LogLevel) > 0 {
		if l, ok := c.LogLevel["all"]; ok && l != defaultLevel {
			err := log.SetAllLogLevel(l)
			if err != nil {
				return err
			}
			delete(c.LogLevel, "all")
		}
		for sub, lvl := range c.LogLevel {
			err := log.SetLogLevel(sub, lvl)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type P2P struct {
	Identity Identity `yaml:"Identity"`
	// libp2p multi address
	Addrs []string `yaml:"Addrs"`

	BootstrapPeers []string `yaml:"BootstrapPeers"`

	BandWidthReporter BandWidthReporter `yaml:"BandWidthReporter"`

	EnableAutoRelay bool `yaml:"EnableAutoRelay"`

	AutoNATService bool `yaml:"AutoNATService"`

	DHT DHT `yaml:"DHT"`
}

type Proxy struct {
	Protocols []ProxyProtocol `yaml:"Protocols"`

	ServiceAdvertiseInterval time.Duration `yaml:"ServiceAdvertiseInterval"`
}

type Endpoint struct {
	ProxyProtocols []ProxyProtocol `yaml:"ProxyProtocols"`

	ServiceDiscoveryInterval time.Duration `yaml:"ServiceDiscoveryInterval"`

	Balancer string `yaml:"Balancer"`
}

type Identity struct {
	PrivKey string `yaml:"PrivKey"`

	ObservedAddrActivationThresh int `yaml:"ObservedAddrActivationThresh"`
}

type BandWidthReporter struct {
	Enable bool `yaml:"Enable"`

	Interval time.Duration `yaml:"Interval"`
}

type DHT struct {
	Client bool `yaml:"Client"`
}

type ProxyProtocol struct {
	Protocol string                 `yaml:"Protocol"`
	Listen   string                 `yaml:"Listen"`
	Config   map[string]interface{} `yaml:"Config"`
}
