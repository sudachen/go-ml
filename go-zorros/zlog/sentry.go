package zlog

import (
	"github.com/getsentry/sentry-go"
	"time"
)

const flashTimeout = 3 * time.Second

var sentryFatalLog = &snio{sentry.LevelFatal}
var sentryErrorLog = &snio{sentry.LevelError}
var sentryWarnLog = &snio{sentry.LevelWarning}
var sentryInfoLog = &snio{sentry.LevelInfo}

type snio struct {
	level sentry.Level
}

func sentryOutput(p []byte, level sentry.Level) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
		sentry.CaptureMessage(string(p))
	})
	if level == sentry.LevelFatal {
		sentry.Flush(flashTimeout)
	}
}

func (sn *snio) Write(p []byte) (n int, err error) {
	if sentry.CurrentHub().Client() != nil {
		sentryOutput(p, sn.level)
	}
	return 0, nil
}

func (sn *snio) Close() error {
	sentry.Flush(flashTimeout)
	return nil
}
