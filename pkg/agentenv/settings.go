package agentenv

import (
	"fmt"
	"time"
)

type Settings struct {
	OutputPath string

	AnalysisProvider string
	AnalysisModel    string
	AnalysisTimeout  time.Duration

	GenerateClean bool
}

func (o *Settings) Validate() error {
	if o == nil {
		return fmt.Errorf("options cannot be nil")
	}
	return nil
}
