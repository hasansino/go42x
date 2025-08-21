package cmdutil

import "github.com/spf13/pflag"

type Options struct {
	Quiet bool
}

func (o *Options) BindFlags(f *pflag.FlagSet) {
	f.BoolVarP(&o.Quiet, "quiet", "q", false, "Disable all output except errors")
}
