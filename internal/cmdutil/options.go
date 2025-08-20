package cmdutil

import "github.com/spf13/pflag"

type Options struct {
	Debug bool
	Quiet bool
}

func (o *Options) BindFlags(f *pflag.FlagSet) {
	f.BoolVarP(&o.Debug, "debug", "d", false, "Enable debug output")
	f.BoolVarP(&o.Quiet, "quiet", "q", false, "Disable all output except errors")
}
