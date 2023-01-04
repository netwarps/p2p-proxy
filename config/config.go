package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/diandianl/p2p-proxy/log"
	"github.com/diandianl/p2p-proxy/metadata"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	DefaultConfigPath = "~/p2p-proxy.yaml"
)

var InvalidErr = errors.New("config invalid or not checked")

func init() {
	if peers, ok := os.LookupEnv("P2P_PROXY_DEFAULT_BOOT_PEERS"); ok {
		Default.P2P.BootstrapPeers = strings.Fields(peers)
	}
}

var Default = &Config{
	P2P: P2P{
		Addrs: []string{
			"/ip4/0.0.0.0/udp/8888/quic",
		},
		BootstrapPeers: []string{},
	},
	Logging: Logging{
		File:   "~/p2p-proxy.log",
		Format: "color",
		Level: map[string]string{
			"all": "info",
		},
	},
	Version:    metadata.Version,
	ServiceTag: "p2p-proxy/0.0.1",
	Proxy: Proxy{
		Protocols: []Protocol{
			{
				Protocol: "/p2p-proxy/http/0.1.0",
				Config:   map[string]interface{}{},
			},
			{
				Protocol: "/p2p-proxy/shadowsocks/0.1.0",
				Config:   map[string]interface{}{},
			},
		},
		ServiceAdvertiseInterval: time.Hour,
	},
	Endpoint: Endpoint{
		ProxyProtocols: []ProxyProtocol{
			{
				Protocol: "/p2p-proxy/http/0.1.0",
				Listen:   "127.0.0.1:8010",
			},
			{
				Protocol: "/p2p-proxy/shadowsocks/0.1.0",
				Listen:   "127.0.0.1:8020",
			},
		},

		Balancer: "round_robin",
	},
	Interactive: false,
}

func LoadOrInitializeIfNotPresent(configPath string) (cfg *Config, cfgFile string, err error) {
	cfgFile = configPath
	if len(cfgFile) == 0 {
		cfgFile = DefaultConfigPath
	}

	cfgFile, err = homedir.Expand(filepath.Clean(cfgFile))
	if err != nil {
		return
	}
	_, err = os.Stat(cfgFile)
	if err != nil {
		if os.IsNotExist(err) {
			cfg, err = Initialize(cfgFile)
			if err != nil {
				return nil, cfgFile, err
			}
			return cfg, cfgFile, err
		}
		return
	}
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return
	}
	cfg = new(Config)
	err = viper.Unmarshal(cfg)
	if err != nil {
		return
	}

	if viper.GetString("Version") == "v0.0.1" {
		cfg.P2P.Identity.PrivKey = viper.GetString("Identity.PrivKey")
		cfg.P2P.Addrs = viper.GetStringSlice("P2P.Addr")

		cfg.Endpoint = Default.Endpoint
		cfg.Proxy = Default.Proxy
		cfg.ServiceTag = Default.ServiceTag
		cfg.Version = metadata.Version
		cfg, err = writeConfig(cfgFile, cfg)
		if err != nil {
			return nil, cfgFile, err
		}
	}
	return cfg, cfgFile, nil
}

func Initialize(cfgPath string) (*Config, error) {
	var cfg Config = *Default

	if runtime.GOOS == "windows" {
		// FIXME  zap log output file, using url.Parse to parse file path, not work on windows
		cfg.Logging.File = ""
	}

	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return nil, err
	}
	privKey, err := crypto.MarshalPrivateKey(priv)

	if err != nil {
		return nil, err
	}
	cfg.P2P.Identity.PrivKey = base64.StdEncoding.EncodeToString(privKey)

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

	Logging Logging `yaml:"Logging"`

	ServiceTag string `yaml:"ServiceTag"`

	P2P P2P `yaml:"P2P"`

	Proxy Proxy `yaml:"Proxy"`

	Endpoint Endpoint `yaml:"Endpoint"`

	Interactive bool `yaml:"Interactive"`

	valid      bool `yaml:"-"`
	work4proxy bool `yaml:"-"`
}

func (c *Config) Validate(proxy bool) error {
	if c.valid {
		return nil
	}
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
	c.work4proxy = proxy
	return nil
}

func (c *Config) Work4Proxy() bool {
	return c.work4proxy
}

func (c *Config) SetupLogging(defaultLevel string) (err error) {
	if !c.valid {
		return InvalidErr
	}

	if len(defaultLevel) == 0 {
		defaultLevel = c.Logging.Level["all"]
	}
	delete(c.Logging.Level, "all")

	var logFile string
	if len(c.Logging.File) > 0 {
		logFile, err = homedir.Expand(filepath.Clean(c.Logging.File))
		if err != nil {
			return err
		}
	}

	err = log.SetupLogging(logFile, c.Logging.Format, defaultLevel)
	if err != nil {
		return err
	}

	if len(c.Logging.Level) > 0 {
		for sub, lvl := range c.Logging.Level {
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
}

type Proxy struct {
	Protocols []Protocol `yaml:"Protocols"`

	ServiceAdvertiseInterval time.Duration `yaml:"ServiceAdvertiseInterval"`
}

type Endpoint struct {
	ProxyProtocols []ProxyProtocol `yaml:"ProxyProtocols"`

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

type ProxyProtocol struct {
	Protocol string `yaml:"Protocol"`
	Listen   string `yaml:"Listen"`
	// Config   map[string]interface{} `yaml:"Config"`
}

type Protocol struct {
	Protocol string                 `yaml:"Protocol"`
	Config   map[string]interface{} `yaml:"Config"`
}

type Logging struct {
	File string `yaml:"File"`
	// json console nocolor, default nocolor
	Format string            `yaml:"Format"`
	Level  map[string]string `yaml:"Level"`
}
