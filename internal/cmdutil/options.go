package cmdutil

import "github.com/spf13/pflag"

type Options struct {
	Quiet    bool
	LogLevel string
}

func (o *Options) BindFlags(f *pflag.FlagSet) {
	f.BoolVarP(&o.Quiet, "quiet", "q", false, "Disable all output except errors")
	f.StringVar(&o.LogLevel, "log-level", "info", "Logging level (debug, info, warn, error)")
}
