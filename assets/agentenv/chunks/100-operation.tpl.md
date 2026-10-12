### Operations

- Ignore anything between `[IGNORE]` and `[/IGNORE]` tags in prompts.
- Never modify root README.md without explicit instructions.

### Communication

- **Clarity**: Be concise and direct in explanations.
- **Context**: Provide reasoning for technical decisions.
- **Progress**: Communicate status during long-running operations.

### Intermediate storage

Use `{{ .context.TempDir }}` directory for intermediate storage of files, artifacts, reports, and other outputs generated during
the operation.

### Side effect protection in interactive sessions

- **NEVER** commit changes to the repository unless explicitly instructed to do so.
- **ALWAYS** ask for confirmation before executing cli tools commands which will produce side effects,
  for example: `git push`, `git commit`, `docker push`, `kubectl apply`.
- **NEVER** produce side effects in the repository or external systems when user is asking a question, opinion or for advice.
