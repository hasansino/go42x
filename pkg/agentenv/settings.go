package agentenv

import (
	"fmt"
	"time"
)

type Settings struct {
	ConfigPath string
	OutputPath string

	AnalysisProvider string
	AnalysisTimeout  time.Duration
}

func (o *Settings) Validate() error {
	if o == nil {
		return fmt.Errorf("options cannot be nil")
	}
	return nil
}
