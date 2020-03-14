package log

import (
	"fmt"
	"github.com/ipfs/go-log/v2"
	"os"
	"path/filepath"
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

	envs := []struct {
		key string
		val string
	}{
		{"GOLOG_LOG_FMT", format},
		{"GOLOG_FILE", file},
		{"GOLOG_LOG_LEVEL", level},
	}

	setup := false
	for _, e := range envs {
		if len(e.val) > 0 {
			setup = true
			err := os.Setenv(e.key, e.val)
			if err != nil {
				return err
			}
		}
	}
	if len(file) > 0 {
		err := os.MkdirAll(filepath.Dir(file), 0755)
		if err != nil {
			return err
		}
	}
	if setup {
		log.SetupLogging()
	}
	return nil
}
