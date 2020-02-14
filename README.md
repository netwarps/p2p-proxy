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
start a proxy server peer

Usage:
  p2p-proxy proxy [flags]

Flags:
  -h, --help   help for proxy

Global Flags:
  -c, --config string      config file (default is $HOME/.p2p-proxy.yaml)
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
A p2p based http(s) proxy

Usage:
  p2p-proxy [flags]
  p2p-proxy [command]

Available Commands:
  help        Help about any command
  init        generate and write default config
  proxy       start a proxy server peer

Flags:
  -c, --config string      config file (default is $HOME/.p2p-proxy.yaml)
  -h, --help               help for p2p-proxy
      --http string        local http(s) proxy agent listen address
      --p2p-addr strings   peer listen addr(s)
  -p, --proxy string       proxy server address
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
# p2p 节点id
identity:
  privkey: CAASqQ .... /OsP53EAmA==
# p2p 监听地址信息
p2p:
  addr:
  - /ip4/0.0.0.0/tcp/8888
# 本地端配置
endpoint:
# http(s) 代理 本地监听地址端口
  http: 127.0.0.1:8010
# 远端代理服务器地址
  proxy: "/ip4/127.0.0.1/tcp/8888/ipfs/QmXktjtfrwwjPYowB9DV7qdViLnjAKHbYgGA4oSjuwYAAY"
# 代理服务器配置
proxy:
  auth:
#   基本认证
    basic:
      realm: my_realm
      users:
#       用户名密码对
        foo: bar
```