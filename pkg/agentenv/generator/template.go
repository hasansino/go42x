package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

const (
	chunksPlaceholder    = "{{.chunks}}"
	modesPlaceholder     = "{{.modes}}"
	workflowsPlaceholder = "{{.workflows}}"
)

type TemplateEngine struct {
	baseDir string
	funcs   template.FuncMap
}

func newTemplateEngine(baseDir string) *TemplateEngine {
	return &TemplateEngine{
		baseDir: baseDir,
		funcs:   defaultFuncs(),
	}
}

func defaultFuncs() template.FuncMap {
	return template.FuncMap{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"trim":  strings.TrimSpace,
		"join":  strings.Join,
	}
}

func (e *TemplateEngine) Process(content string, ctx *Context) (string, error) {
	tmpl, err := template.New("main").Funcs(e.funcs).Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx.ToMap()); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (e *TemplateEngine) MergeStrings(items []string) string {
	var result strings.Builder

	for i, item := range items {
		if i > 0 {
			result.WriteString("\n\n")
		}
		result.WriteString(strings.TrimSpace(item))
	}

	return result.String()
}
