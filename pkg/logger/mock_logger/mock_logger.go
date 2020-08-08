package mock_logger

import (
	"time"

	logger2 "github.com/int128/kauthproxy/pkg/logger"
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
	logf(l.t, "", format, args)
}

func (l *Logger) V(level int) logger2.Verbose {
	return &Verbose{l.t}
}

type Verbose struct {
	t testingLogf
}

func (v *Verbose) Infof(format string, args ...interface{}) {
	logf(v.t, "I]", format, args)
}

func logf(t testingLogf, level, format string, args []interface{}) {
	t.Logf("%s %2s "+format, append([]interface{}{
		time.Now().Format("15:04:05.000"),
		level,
	}, args...)...)
}
