package collector

import (
	"context"
	"os"
	"path/filepath"
)

const ConventionsCollectorName = "conventions"

const conventionsFileName = "CONVENTIONS.md"

// ConventionsCollector collects data from CONVENTIONS.md
type ConventionsCollector struct {
	BaseCollector
	rootDir string
}

func NewConventionsCollector(rootDir string) *ConventionsCollector {
	return &ConventionsCollector{
		BaseCollector: NewBaseCollector(ConventionsCollectorName, 50),
		rootDir:       rootDir,
	}
}

func (c *ConventionsCollector) Collect(_ context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	analysisFilePath := filepath.Join(c.rootDir, conventionsFileName)

	data, err := os.ReadFile(analysisFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			result["content"] = ""
			return result, nil
		}
		return nil, err
	}

	result["content"] = string(data)

	return result, nil
}
