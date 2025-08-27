package kwb

import (
	"fmt"
	"os"
	"time"
)

type Settings struct {
	RootPath  string // Directory to index
	IndexPath string // Path to store the index

	ExtraExtensions []string
	ExcludeDirs     []string
	MaxFileSize     int

	SearchTimeout   time.Duration
	SearchLimit     int
	SearchShowScore bool
}

func (s *Settings) Validate() error {
	if s == nil {
		return fmt.Errorf("settings cannot be nil")
	}
	return nil
}

func (s *Settings) IndexExists() bool {
	if _, err := os.Stat(s.IndexPath); os.IsNotExist(err) {
		return false
	}
	return true
}
