package collector

import (
	"context"
	"os"
	"path/filepath"
)

const analysisFileName = "analysis.gen.md"

// AnalysisCollector collects data from analysis.gen.md
type AnalysisCollector struct {
	BaseCollector
	templateDir string
}

func NewAnalysisCollector(templateDir string) *AnalysisCollector {
	return &AnalysisCollector{
		BaseCollector: NewBaseCollector(
			"analysis",
			50,
		),
		templateDir: templateDir,
	}
}

func (c *AnalysisCollector) Collect(_ context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	analysisFilePath := filepath.Join(c.templateDir, analysisFileName)

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
