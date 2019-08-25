package mock_logger

import (
	"github.com/int128/kauthproxy/pkg/logger"
	"github.com/spf13/pflag"
)

type testingLogf interface {
	Logf(format string, args ...interface{})
}

func New(t testingLogf) *Logger {
	return &Logger{t}
}

type Logger struct {
	t testingLogf
}

func (l *Logger) AddFlags(f *pflag.FlagSet) {
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.t.Logf(format, args...)
}

func (l *Logger) V(level int) logger.Verbose {
	return &Verbose{l.t}
}

type Verbose struct {
	t testingLogf
}

func (v *Verbose) Infof(format string, args ...interface{}) {
	v.t.Logf("I] "+format, args...)
}
