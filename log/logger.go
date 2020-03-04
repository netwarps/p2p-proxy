package log

import (
	"fmt"
	"github.com/ipfs/go-log/v2"
)

const LogSystem = "p2p-proxy"

type Logger = log.StandardLogger

func NewLogger() Logger {
	return log.Logger(LogSystem)
}

func NewSubLogger(sub string) Logger {
	return log.Logger(fmt.Sprintf("%s/%s", LogSystem, sub))
}

func SetAllLogLevel(level string) error {
	return log.SetLogLevel("*", level)
}

func SetLogLevel(name, level string) error {
	return log.SetLogLevel(name, level)
}
