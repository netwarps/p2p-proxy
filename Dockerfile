FROM golang:1.19 as builder

ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.com.cn,direct

COPY . /source/
RUN cd /source/ && \
	go mod tidy && \
	GOOS=linux CGO_ENABLED=1 go build -o p2p-proxy

FROM golang:1.19 as prod
USER root
RUN mkdir /config
COPY --from=0 ./source/p2p-proxy  /
ENV TZ Asia/Shanghai

EXPOSE 8888
EXPOSE 8020

RUN  chmod +x /p2p-proxy
WORKDIR /
ENTRYPOINT ["/p2p-proxy"]

