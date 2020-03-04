package http

import (
	"github.com/diandianl/p2p-proxy/log"

	"github.com/elazarl/goproxy"
)

func setLogger(proxy *goproxy.ProxyHttpServer, logger log.Logger) {
	logf := logger.Infof
	if proxy.Verbose {
		logf = logger.Debugf
	}
	proxy.Logger = &loggerAdapter{logf}
}

type loggerAdapter struct {
	logf func(string, ...interface{})
}

func (a *loggerAdapter) Printf(format string, v ...interface{}) {
	a.logf(format, v...)
}
