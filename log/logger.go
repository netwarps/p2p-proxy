package log

import (
	"fmt"
	"github.com/ipfs/go-log/v2"
)

const LogSystem = "p2p-proxy"

type Logger interface {
	log.StandardLogger
	Sync() error
}

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

func SetupLogging(file string, format string, level string) error {
	_, err := log.LevelFromString(level)
	if err != nil {
		return err
	}

	cfg := log.Config{
		Format: log.ColorizedOutput,
		//File:   file,
		Level:  log.LevelInfo,
		Stdout: true,
	}
	switch format {
	case "nocolor":
		cfg.Format = log.PlaintextOutput
	case "json":
		cfg.Format = log.JSONOutput
	}

	log.SetupLogging(cfg)
	return nil
}
