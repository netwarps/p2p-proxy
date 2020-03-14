# p2p-proxy

`p2p-proxy` 是一个命令行式的，基于[`libp2p`](https://github.com/libp2p/go-libp2p)和[`goproxy`](github.com/elazarl/goproxy)实现的`http(s)`代理工具。

## 实现原理
```
+----------------------+                          +-----------------------------------+
|      Endpoint        |                          |              Proxy                |
| +-----+      +-----+ |   /proxy-example/0.0.1   | +--------------+     +----------+ |
| | TCP | <--> | p2p | |      /secio/1.0.0        | | p2p|gostream + <-> | goproxy  | |
| +-----+      +-----+ | <----------------------> | +--------------+     +----------+ |
+----------------------+                          +-----------------------------------+
```

## 运行方式
`p2p-proxy`为命令行程序，启动分为远端和本地端。

### 远端
远端为实际代理服务器，帮助信息如下：
```
./p2p-proxy proxy -h
Start a proxy server peer

Usage:
  p2p-proxy proxy [flags]

Flags:
  -h, --help   help for proxy

Global Flags:
  -c, --config string      config file (default is $HOME/.p2p-proxy.yaml)
      --log-level string   set logging level (default "INFO")
      --p2p-addr strings   peer listen addr(s)
```
启动示例：
```shell script
# ./p2p-proxy proxy
Using config file: /Users/someuser/.p2p-proxy.yaml
Proxy server is ready
libp2p-peer addresses:
/ip4/127.0.0.1/tcp/8888/ipfs/QmXktjtfrwwjPYowB9DV7qdViLnjAKHbYgGA4oSjuwYAAY
```

### 本地端
本地端为代理服务的入口，目前只支持代理`http(s)`，命令帮助如下：
```
# ./p2p-proxy -h
A http(s) proxy based on P2P

Usage:
  p2p-proxy [flags]
  p2p-proxy [command]

Available Commands:
  help        Help about any command
  init        Generate and write default config
  proxy       Start a proxy server peer

Flags:
  -c, --config string      config file (default is $HOME/.p2p-proxy.yaml)
  -h, --help               help for p2p-proxy
      --http string        local http(s) proxy agent listen address
      --log-level string   set logging level (default "INFO")
      --p2p-addr strings   peer listen addr(s)
  -p, --proxy string       proxy server address

Use "p2p-proxy [command] --help" for more information about a command.
```
启动实例：
```shell script
./p2p-proxy
Using config file: /Users/someuser/.p2p-proxy.yaml
proxy listening on  127.0.0.1:8010
```

## 配置文件说明
如果不指定，默认使用`$HOME/.p2p-proxy.yaml`。程序首次启动时是会自动创建配置文件，并生成节点id等信息写入配置文件。

配置文件说明：
```yaml
# 配置版本
Version: v0.0.2
# 配置日志级别
Logging: 
  File: ~/p2p-proxy.log
  Format: console
  Level:
    all: info
    p2p-proxy: debug
# 在dht网络暴露/查找服务的标签
ServiceTag: p2p-proxy/0.0.1
# p2p 网络配置
P2P:
  Identity:
    # 私钥
    PrivKey: CAASpe...k15GFrqg==
    # 外部观察到的地址激活阈值，0表示使用libp2p默认（4），即有4个外表节点确认同一外部访问地址则讲该地址添加到本节点地址列表
    ObservedAddrActivationThresh: 0
# 本节点监听地址
  Addrs:
  - /ip4/0.0.0.0/tcp/8010
# boot 节点
  BootstrapPeers:
  - /ip4/127.0.0.1/tcp/8888/ipfs/Qm...MW
  - /dns4/proxy.server.com/tcp/8888/ipfs/QmW...yMW
# 带宽流浪统计
  BandWidthReporter:
    Enable: false
    Interval: 0s
# 启用自动中继
  EnableAutoRelay: false
# 启用NAT 服务端
  AutoNATService: false
# DHT 配置
  DHT:
    Client: false
# 代理服务设置
Proxy:
  # 支持的协议列表
  Protocols:
  - Protocol: /p2p-proxy/http/0.0.1
    Config: {}
  ServiceAdvertiseInterval: 1h0m0s
# 本地端配置
Endpoint:
  # 本地端支持（监听）的协议，由远端提供支持
  ProxyProtocols:
  - Protocol: /p2p-proxy/http/0.0.1
    # 协议监听地址
    Listen: 127.0.0.1:8010
  # 代理服务发现时间间隔
  ServiceDiscoveryInterval: 1h0m0s
  # 代理服务节点均衡策略
  Balancer: round_robin
# 开启交互模式，提供 cli 命令查看内部信息
Interactive: false
```