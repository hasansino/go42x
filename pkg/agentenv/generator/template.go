package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

const (
	chunksPlaceholder    = "{{ .chunks }}"
	modesPlaceholder     = "{{ .modes }}"
	workflowsPlaceholder = "{{ .workflows }}"
)

type TemplateEngine struct {
	baseDir   string
	functions template.FuncMap
}

func newTemplateEngine(baseDir string) *TemplateEngine {
	return &TemplateEngine{
		baseDir:   baseDir,
		functions: defaultFuncs(),
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

func (e *TemplateEngine) Process(content string, ctxData map[string]interface{}) (string, error) {
	tmpl, err := template.New("main").Funcs(e.functions).Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctxData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (e *TemplateEngine) InjectChunks(content string, chunks string) string {
	return e.inject(content, chunks, chunksPlaceholder)
}

func (e *TemplateEngine) InjectModes(content string, modes string) string {
	return e.inject(content, modes, modesPlaceholder)
}

func (e *TemplateEngine) InjectWorkflows(content string, workflows string) string {
	return e.inject(content, workflows, workflowsPlaceholder)
}

func (e *TemplateEngine) inject(content string, payload string, placeholder string) string {
	return strings.Replace(content, placeholder, payload, 1)
}
