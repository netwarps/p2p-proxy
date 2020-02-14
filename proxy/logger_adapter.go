package proxy

import "github.com/elazarl/goproxy"

func NewLoggerAdapter(proxy *goproxy.ProxyHttpServer) *loggerAdapter {
	return &loggerAdapter{proxy}
}
type loggerAdapter struct {
	proxy *goproxy.ProxyHttpServer
}

func (l *loggerAdapter) Printf(format string, v ...interface{}) {
	if l.proxy.Verbose {
		log.Debugf(format, v...)
	} else {
		log.Infof(format, v...)
	}
}
