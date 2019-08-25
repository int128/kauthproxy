// Package logger provides logging facility.
package logger

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/wire"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

var Set = wire.NewSet(
	wire.Bind(new(Interface), new(*Logger)),
	wire.Struct(new(Logger)),
)

type Interface interface {
	AddFlags(f *pflag.FlagSet)
	Printf(format string, args ...interface{})
	V(level int) Verbose
}

type Verbose interface {
	Infof(format string, args ...interface{})
}

// Logger provides logging facility using klog.
type Logger struct{}

// AddFlags adds the flags such as -v.
func (*Logger) AddFlags(f *pflag.FlagSet) {
	gf := flag.NewFlagSet("", flag.ContinueOnError)
	klog.InitFlags(gf)
	f.AddGoFlagSet(gf)
}

// Printf writes the message to stderr.
func (*Logger) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	_, _ = fmt.Fprintln(os.Stderr, "")
}

// V returns a logger enabled only if the level is enabled.
func (*Logger) V(level int) Verbose {
	return klog.V(klog.Level(level))
}
