# Go Code Style

## Project defaults

Follow explicit user instructions and project conventions before these defaults
or upstream recommendations. Use the project's prescribed formatters and checks.

- Write function bodies and control flow on multiple lines, including functions
  that only return a value. Avoid one-line functions.
- Use blank lines to separate logical steps. Keep related statements together
  and preserve the surrounding layout.
- Group imports with the standard library first, then external dependencies,
  then project packages. Separate groups with a blank line.
- Put unit tests beside the source they test: `foo.go` uses `foo_test.go`.
  Put integration and end-to-end tests in `tests/integration/` and `tests/e2e/`.

## Read detailed conventions on demand

Before editing or reviewing Go code, consult the sections relevant to the task
using the table below. Paths are relative to the project root. These Markdown
snapshots are installed locally by go42x; no network access is needed.

Search the selected file for a heading or symbol with `rg` or an available local
search tool, then read that section with a file-reading tool. Follow relevant
cross-references when more detail is needed. Read only the necessary sections;
do not load every reference into the initial context.

| Task or question                                 | Local reference                               | Sections to find                                                          |
|--------------------------------------------------|-----------------------------------------------|---------------------------------------------------------------------------|
| Common review issues: names, imports, APIs       | `.go42x/upstream/go/code-review-comments.md`  | Imports, Initialisms, Interfaces, Receiver Type, Contexts, Handle Errors  |
| Choosing a Google style document                 | `.go42x/upstream/google-go/index.md`          | About, Documents                                                          |
| Readability, simplicity, formatting, consistency | `.go42x/upstream/google-go/guide.md`          | Style principles, Core guidelines, Local consistency                      |
| Detailed rules for code, comments, errors, tests | `.go42x/upstream/google-go/decisions.md`      | Naming, Commentary, Imports, Errors, Language, Contexts, Test structure   |
| Practical examples and design tradeoffs          | `.go42x/upstream/google-go/best-practices.md` | Error handling, Documentation, Function argument lists, Tests, Interfaces |

Use Go Code Review Comments for common Go conventions, then the relevant Google
documents for detail. Google's best practices are recommendations; apply them
with judgment. Upstream examples do not mandate this project's architecture,
dependencies, or tooling.
