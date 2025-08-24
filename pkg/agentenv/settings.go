package agentenv

import "fmt"

type Settings struct {
	ConfigPath string
	OutputPath string
}

func (o *Settings) Validate() error {
	if o == nil {
		return fmt.Errorf("options cannot be nil")
	}
	return nil
}
