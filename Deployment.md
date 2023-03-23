## 准备工作
- 安装docker
- 两台服务器，一台位于中国境内作为proxy stub，一个境外作为proxy server， shadowsocks客户端连接stub，stub链接server，server连接境外网站
## 部署server
1. 创建配置文件目录并生成配置
```
mkdir /root/p2p-proxy-config
docker run --rm -it -v /root/p2p-proxy-config:/config registry.paradeum.com/netwarps/p2p-proxy:v0.0.1 -c /config/p2p-proxy-server.yaml
```
Ctrl+C终止容器, 查看/root/p2p-proxy-config目录下生成的配置文件p2p-proxy-server.yaml

2. 修改配置
- 默认使用quic协议，监听端口8888
- 删掉Proxy中的http协议，删除Endpoint，配置shadowsocks的ciper和password
```
Version: dev-build
Logging:
  File: ~/p2p-proxy.log
  Format: color
  Level:
    all: info
ServiceTag: p2p-proxy/0.0.1
P2P:
  Identity:
    PrivKey: CAASqQkwggSlAgEAAoIBAQCjC0trGQD3eel5jnkqVtmHm4c9EXhDuQcyESai6RtAaZ3HCssH8Xx7bqY+D/GC6PSF15jDrwEdsMRxdJaliwKVAZ6N2jQ+9KHY5OPdqYGUXFo1Zwxi7OfvkdhVkVlgBWPyna2tUfuyONMwkLZDDmRH7LpfLRMCtVdU5IviINP4eTbnmoIB+r2wLpuS2tPFz3wee2vwwm1ml4n73x/+DfNt3cYRebB75+n9dTTvrGiXbKsLLl6u6b04kOq1onylUGVs/PYAEQbSUN3e4P/RdmvxhCifDuWooLaGapLTbLAHdcOTD0WUYkpGls7g2doSZsR9c6wq8au9j9eTZEw4zfGtAgMBAAECggEAfsTtYtwSEFlN2yGXu//DKtkWkbjflWhr29XSAKDWe4KjFnuh2Q8+BorF30NuOKcAWICFWsDbUUZ7tus7poMrAsg7i3e5X6m9nXJ6aYK+KaiUyyjQTKp+u5reZcmZgDswtxc6TqSL2sqsCfq6e/DYr8O0NQRK37Q3rt30lWGI7ouzGvrzzhhPsX046y3UdZJA9bTA+BQvNlqnKwEp+mT+7fP95Y8k9blUA5b1p1+dahX/vNoGnwmeEbhWX9LODc3kyOTVG5GEQZpedvEY5UdHdedzW7P1T/p1KThAoHjYP1/VSlpoQz/7CjSFKZHy5J5i32YDuSJx+5rLy+08D7fyUQKBgQDKrWNZNbQ5OvjXql5zOI+s8D/TNbKlojDp+PXD9Xfd7Ch1AFWE44DTi/Ei2tRNo0S3/8lB1BmwK0cKhIeYFkcOJyBTR0QgcVuXXQ89PfODbacoAYNn6tMCmP5Y9x7TkZbQQhhCg6WDKS6ECztc1DLrBnGQ3zKDcQfKGbBKpZ0gcwKBgQDN8I3bzkjRxv7Fac7z1i7qh/jOD4nYeV++oQX9tMrEuho34RElpIJJ4kxrUKSOQTTIC5tvKWLOCO+QnJyKkRtLJCIOdAoh8nAzbfyveSo4z9StKXR/G7ZTY5QqyPwWfvQrEi/uHgT9kDygnN+ETcU0aMvKy+gY0J31xNDiQs69XwKBgQC/9JPhjAGTKo1ABTXLPsik7C4m5fa69PAKySZLYBMU9nQizBwy7h23PhU2A7eLiJSvB+0fEbj6pyJzja22l3LYrqno9dhKOdKbeyHRyPj3g0ULmNNR+o+7KBfNPs/NZVhHCjJb3L9HiBtsKA8jDj7jZYjtwtbespDEEqxrJou4jQKBgQCtUNzidxpbygCSPfkYx1HWubZQHU2ibIuCkFvNaAEaTZFRI85dgrTP3273BfhnbEMydGpMxGTOB0Eu0E8CYxq4Q2GSDmCUr0d0UQVO3EcHZwmS7geIDdeFGJIS6/EUMaXmNbk2yfbjOyd6+Gs4God0ExonwzHC6Jd3xjsRoK4DOwKBgQC8u92kL0He+dE5eQqr9wPj0MM5h208y05NxSJHHYcwXdwAAkKM9Mxg7XSCRAu8GmBgaDnQVT8J5ukJOMW9Rv9ZWgzRbOkNmMIzUhMlWzY3sTZ/hFJQkML2q2s6nHr9FEWXK5ePLpy2gVRc2nT6t9J99FCLnzvZUIfbYMtmZvg8BQ==
    ObservedAddrActivationThresh: 0
  Addrs:
  - /ip4/0.0.0.0/udp/8888/quic #修改
  BootstrapPeers: []
  BandWidthReporter:
    Enable: false
    Interval: 0s
Proxy:
  Protocols:
  - Protocol: /p2p-proxy/shadowsocks/0.1.0
    Config: 
      ciper: AES-128-GCM  #修改shadowsocks加密方式
      password: 666999  #修改shadowsocks密码
  ServiceAdvertiseInterval: 1h0m0s
Interactive: false
```
3. 运行server
```
docker run -d --name p2p-proxy-server -p 8888:8888 -v /root/p2p-proxy-config/p2p-proxy-server.yaml:/root/p2p-proxy.yaml registry.paradeum.com/netwarps/p2p-proxy:v0.0.1  proxy
```
4. 获取peer地址
```
docker logs -f  p2p-proxy-server
```
查看日志,可发现类似下面的日志
```
2023-03-23T17:54:30.030+0800	INFO	p2p-proxy/p2p	p2p/libp2p.go:46	P2P [/ipfs/Qmc57rUkvVX8UxUxJDpP5uk2esjNLZKydbSKmFYjaoBf6W] working addrs: [/ip4/172.30.0.228/udp/8888/quic /ip4/127.0.0.1/udp/8888/quic]
```
拼接起来,将ip替换成你的公网ip, stub的BootstrapPeers配置该地址
```
/ip4/XX.XX.XX.XX/udp/8888/quic/ipfs/Qmc57rUkvVX8UxUxJDpP5uk2esjNLZKydbSKmFYjaoBf6W
```
## 部署stub
1. 创建配置文件目录,生成配置
```
mkdir /root/p2p-proxy-config
docker run --rm -it -v /root/p2p-proxy-config:/config registry.paradeum.com/netwarps/p2p-proxy:v0.0.1 -c /config/p2p-proxy-stub.yaml
```
Ctrl+C终止容器, 查看/root/p2p-proxy-config目录下生成的配置文件p2p-proxy-stub.yaml
2. 修改配置
- 默认使用quic协议,监听端口8888
- 修改BootstrapPeers,地址为p2p-proxy-server的peer地址(部署server的第四步)
- 删除Proxy保留Endpoint,配置shadowsocks监听端口,默认127.0.0:8020,修改为0.0.0.0:8020
```
Version: dev-build
Logging:
  File: ~/p2p-proxy.log
  Format: color
  Level:
    all: info
ServiceTag: p2p-proxy/0.0.1
P2P:
  Identity:
    PrivKey: CAASpwkwggSjAgEAAoIBAQDKnaYyJ8JmanNUyrf4qogIe6ZFml03JKHtt+fsQaHIVp3/UfemaJjW+jlNMu5H0carWi5F2xOvDQW1MNAs7WZtDEq+gU+oW5R7kS0ORn96LVB3eR6cQkTl9sddJTmZU2IUk4bCoEyko5dVwc7JIaNPz/6vGtMmSDSwSO8d8MXcy6X+v8Xe2Z0w6+m4DSoRZMmi125KRpjiPvyptCN4DwAvGGwm6VhTdDtvgLhLdEoWklrZkji46zUDZtiSMqADIQkc36u36Et2mfU1HqD3lgNfSF9XjMcrtvIupDP7srGaQ00p8hamNwo0wyH1k/XO0xfA3tIasWCJ5BP50xX/2qz5AgMBAAECggEAOELqYUb1DidE+yiHST9hIqnjE7S3aZZ8eFv2xH29BLo4iSsjj0vAFQHKY4te6wZvGimia7dXkeYVzahORgttw54EKz4Q9njnlCBN2Ibu4uguTd6OB2nHY+vQlCbABblHpNsKMoT8g0MBxMhaOTNj+8ePuuPB+gFW0BSQgUnYR2SucAm8Jr74ncil9n6DkoJF5+ognkPsx/O2j5bTDWxdYI1dF9zHT8yM/3i0LOvf0cIl9d7af9/PXKEk28rH6M+9JqiTlsBS+SF8rDoU9+nSJir9XBSpXiz9jwlgX6xZRknT+A2FtdfIRnibOzHKZMw3+ebnluum8MoeKp8yHBP6+QKBgQDxwVzz1qozPJeoophPLtZLvwuVqvd7JntKrpRWiGecvlQ/4Sneplq7zTkYxJ0eBGE9yLcBSyGitaQlpM1CcEnoghVYFtgDekLUiq1eiIqZpeJrnRCmZiO0UcXFAu5qawlEE6FldAhvwJ1jhprDzKft8L/moIdfEyXRwW8ceNN2PwKBgQDWjeflaQKMLx0HZB13hquFbWdlmd8RaX2o/NIqyne3cEcPz44GkWN7eDjiqXlU5moQCgIQ0USsNbnmX30AykyelmCB8uGb8rYyyeox4/Zd/fmq0NLgN+kdADF30colGqUAQSTOEki9MmI1yOgRI3jG3FGngW7+nbH6dzmQpom+xwKBgFPtrTd57tyazIve5sGWoQ7q5Dqxf/lhAqyKrzTbZh0kdls28DI7zoQkWw4eM+2X16p7ZA0u6B50sOfgruHB2ea+Qmqyg4uxhkIDYuzOuk9dJ530iTM7gmm3edFLkzmerzjTF9UA02z4katbr58KDcKtMfH/CQAYxahsXwaja8ZBAoGBAM7pzcFFk0pkSUeOeoiB3LpxtuyaBzGAncox//F6jxfedPm/fcXBwsIZQCr/q95/07uiGznix6qYqa6NWj0/28J5XZsVBBTkbmfuqCfzI+6jd3sPpr7LzMnGHO7j6GH+HzBuorMFmRa1F1etaHjWz6xgX3L+dW+h3zmgb2ib422TAoGAPgg2QXjBn/PUfMfkkTk6OjsSE/kroxsQXdKry+MBPKZZPUYOAtgDpb7SwlVRE4NlPzUh9nAtkOhO1ROKrznbjBoIv8ZFEegMsB4Mp+VM4ggid1SD8saRFg7RwEV/7BxHY24ydxvgCCEdKpkAIstXq9il4fTspjldn64XibfiF84=
    ObservedAddrActivationThresh: 0
  Addrs:
  - /ip4/0.0.0.0/udp/8888/quic
  BootstrapPeers: 
  - /ip4/172.30.0.228/udp/8888/quic/ipfs/Qmc57rUkvVX8UxUxJDpP5uk2esjNLZKydbSKmFYjaoBf6W # 修改
  BandWidthReporter:
    Enable: false
    Interval: 0s
Endpoint:
  ProxyProtocols:
  - Protocol: /p2p-proxy/shadowsocks/0.1.0
    Listen: 0.0.0.0:8020  #修改
  Balancer: round_robin
Interactive: false
```
3. 启动stub
```
docker run -d --name p2p-proxy-stub -p 8020:8020 -v /root/p2p-proxy-config/p2p-proxy-stub.yaml:/root/p2p-proxy.yaml registry.paradeum.com/netwarps/p2p-proxy:v0.0.1 
```
## 注意事项
- server和stub不建议部署在境外同一台机器上,否则容易被封
- 如果在桌面电脑上使用，stub可以跑再你的桌面电脑上，可以省掉一台服务器
