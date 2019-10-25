package internal

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	LogFormatLogfmt = "logfmt"
	LogFormatJSON   = "json"

	LogLevelError = "error"
	LogLevelWarn  = "warn"
	LogLevelInfo  = "info"
	LogLevelDebug = "debug"
)

func NewLogger(logLevel, logFormat, name string) log.Logger {
	var (
		logger log.Logger
		lvl    level.Option
	)

	switch logLevel {
	case LogLevelError:
		lvl = level.AllowError()
	case LogLevelWarn:
		lvl = level.AllowWarn()
	case LogLevelInfo:
		lvl = level.AllowInfo()
	case LogLevelDebug:
		lvl = level.AllowDebug()
	default:
		panic("unexpected log level")
	}

	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	if logFormat == LogFormatJSON {
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	}

	logger = level.NewFilter(logger, lvl)
	logger = log.With(logger, "name", name)

	return log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
}
