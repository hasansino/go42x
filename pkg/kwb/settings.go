package kwb

import (
	"fmt"
	"os"
	"time"
)

type Settings struct {
	RootPath  string // Directory to index
	IndexPath string // Path to store the index

	// Indexing options
	ExtraExtensions []string
	ExcludeDirs     []string
	MaxFileSize     int
	BatchSize       int    // Number of documents to index in a batch
	IndexType       string // Index type: "scorch" (default) or "upsidedown"

	// Search options
	SearchTimeout   time.Duration
	SearchLimit     int
	SearchShowScore bool
	SearchFuzziness int    // Fuzzy search distance (0 = exact match, 1-2 = fuzzy)
	HighlightStyle  string // Highlight style: "html" or "ansi"
}

func (s *Settings) Validate() error {
	if s == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	// Apply defaults
	if s.BatchSize <= 0 {
		return fmt.Errorf("batch size must be greater than 0")
	}
	if s.IndexType == "" {
		return fmt.Errorf("index type cannot be empty")
	}
	if s.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be greater than 0")
	}
	if s.SearchLimit <= 0 {
		return fmt.Errorf("search limit must be greater than 0")
	}
	if s.HighlightStyle == "" {
		return fmt.Errorf("highlight style cannot be empty")
	}

	// Validate values
	if s.SearchFuzziness < 0 || s.SearchFuzziness > 2 {
		return fmt.Errorf("search fuzziness must be between 0 and 2")
	}
	if s.IndexType != "scorch" && s.IndexType != "upsidedown" {
		return fmt.Errorf("invalid index type: %s (must be 'scorch' or 'upsidedown')", s.IndexType)
	}

	return nil
}

func (s *Settings) IndexExists() bool {
	if _, err := os.Stat(s.IndexPath); os.IsNotExist(err) {
		return false
	}
	return true
}
