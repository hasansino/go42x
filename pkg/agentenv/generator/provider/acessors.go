package provider

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type TemplateEngineAccessor interface {
	Process(template string, ctxData map[string]interface{}) (string, error)
	InjectChunks(template string, chunks string) string
	InjectModes(template string, modes string) string
	InjectWorkflows(template string, workflows string) string
}
