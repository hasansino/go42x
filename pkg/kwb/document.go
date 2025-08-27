package kwb

import (
	"path/filepath"
	"strings"
)

type Document struct {
	ID      string `json:"id"`
	Path    string `json:"path"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

func GetFileType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return "code"
	case ".md":
		return "documentation"
	case ".yaml", ".yml":
		return "config"
	case ".proto":
		return "proto"
	case ".sql":
		return "sql"
	case ".json":
		return "json"
	case ".toml":
		return "toml"
	case ".mod", ".sum":
		return "module"
	case ".sh":
		return "shell"
	case "":
		base := filepath.Base(path)
		if strings.Contains(base, "Makefile") {
			return "makefile"
		}
		if strings.Contains(base, "Dockerfile") {
			return "dockerfile"
		}
		return "other"
	default:
		return "other"
	}
}
