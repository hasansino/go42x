# Project: {{ .project.Name }}

<context>
  <language>{{ .project.Language }}</language>
  {{ if gt (len .project.Tags) 0 -}}
  <tags>
    {{- range .project.Tags }}
    - {{ . }}
    {{- end }}
  </tags>
  {{- end }}
  {{ if gt (len .project.Metadata) 0 -}}
  <metadata>
    {{- range $key, $value := .project.Metadata }}
    <{{ $key }}>{{ $value }}</{{ $key }}>
    {{- end }}
  </metadata>
  {{- end }}
</context>

{{ .project.Description }}

## Instructions

{{ .chunks }}

## Operational Modes

{{ .modes }}

## Workflows

{{ .workflows }}

{{ if gt (len .analysis) 0 -}}
{{ .analysis }}
{{ end -}}
