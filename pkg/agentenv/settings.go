package agentenv

import (
	"fmt"

	"github.com/go42-dev/go42x/pkg/config"
)

type Settings struct {
	Clean     bool
	Providers []string
}

func (o *Settings) Validate() error {
	if o == nil {
		return fmt.Errorf("settings cannot be nil")
	}
	return config.ValidateProviders(o.Providers)
}
