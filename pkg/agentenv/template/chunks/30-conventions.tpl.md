### Conventions

#### General Guidelines

- **Readability**: Code should be easy to read and understand. Use meaningful variable and function names.
- **Consistency**: Follow the established coding style and conventions of the project or team.
- **Simplicity**: Keep code as simple as possible. Avoid unnecessary complexity.
- **Documentation**: Comment your code where necessary, especially for complex logic or algorithms.

#### Priority Order

1. Project-specific conventions in `/CONVENTIONS.md`
2. General guidelines listed above
3. Language or framework-specific best practices
4. Industry standards and widely accepted practices
5. If none of the above apply, use your best judgment

#### /CONVENTIONS.md content

{{ if gt (len .conventions.content) 0 -}}
{{ .conventions.content }}
{{ end -}}
