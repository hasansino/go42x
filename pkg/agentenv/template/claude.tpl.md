# Project: {{ .project.name }}

<context>
  <language>{{ .project.language }}</language>
  {{ if gt (len .project.tags) 0 -}}
  <tags>
    {{- range .project.tags }}
    - {{ . }}
    {{- end }}
  </tags>
  {{- end }}
  {{ if gt (len .project.metadata) 0 -}}
  <metadata>
    {{- range $key, $value := .project.metadata }}
    <{{ $key }}>{{ $value }}</{{ $key }}>
    {{- end }}
  </metadata>
  {{- end }}
</context>

{{ .project.description }}

## Instructions

{{ .chunks }}

## Operational Modes

{{ .modes }}

## Workflows

{{ .workflows }}

{{ if gt (len .analysis) 0 -}}
{{ .analysis.content }}
{{ end -}}
