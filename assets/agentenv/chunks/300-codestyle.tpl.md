# Go Code Style

## Project defaults

Follow explicit user instructions and project conventions before these defaults
or upstream recommendations. Use the project's prescribed formatters and checks.
These defaults apply to handwritten Go. Generated code follows its generator's
conventions.

### Comments

- Use comment tags such as `@see`, `@todo`, `@fixme`, and `@note` where useful.
- Use `// ---` comments to separate sections in code files.

### Formatting

- Write function bodies and control flow on multiple lines, including functions
  that only return a value. Avoid one-line functions.
- Use blank lines to separate logical steps. Keep related statements together
  and preserve the surrounding layout.
- Write map literals with multiple entries and struct literals with multiple
  fields across multiple lines, including in loop headers, arguments, and return
  statements. Put each entry or field on its own line, use a trailing comma, and
  put the closing brace on a separate line. Empty and single-entry or single-field
  literals may stay on one line.
- End text files with a newline.

### Naming and declarations

| Convention          | Guidance                                               |
| ------------------- | ------------------------------------------------------ |
| Generic values      | Prefer `any` over `interface{}`.                       |
| Context variables   | Name `context.Context` variables `ctx`.                |
| Empty string checks | Prefer `len(s) == 0` over `s == ""` for empty strings. |
| YAML filenames      | Prefer `.yaml` over `.yml`.                            |

- Use named structs and interfaces. Anonymous structs and interfaces, including
  type assertions to anonymous interfaces, are allowed only in tests.
- Define types and constants at package scope. Function-local type and constant
  declarations are allowed only in tests.

### String constants

- Use named string constants for values with shared or stable meaning, such as
  protocol values, configuration keys, or values reused across operations.
- A single-use value can warrant a constant when its name explains that meaning;
  ordinary messages and disposable test data can remain literals.
- Reuse constants from the package that owns the value. Group constants after
  imports and before other declarations. Apply the same judgment in tests.

### Errors

- Wrap errors with `fmt.Errorf` and `%w` when callers need the underlying error.
- Inspect errors with `errors.Is` and `errors.As`, matching identity or type
  rather than message text.
- Return errors from services and libraries. Command entry points decide the
  process exit status.

### Dependency interfaces and boundaries

- Define dependency interfaces in the consuming package with only the methods
  that consumer needs. Keep them beside the consumer or group them in `accessors.go`.
- Prefer names such as `xxxAccessor`, or a simple subsystem name such as `cache`
  or `events`.
- Keep mock-generation directives with the interface declarations and generate
  mocks into the package's `mocks/` directory.
- Entry points compose dependencies and own process lifecycle. Services own
  business operations. Adapters own protocol parsing, conversion, and error
  mapping. Follow the project's existing package layout.
- Parse and validate settings before initializing dependencies.

### Resource ownership

- Give background goroutines and resources explicit owners and shutdown paths.
  Use request contexts for request work and component contexts for workers.
- Bound external calls and retry waits. Keep concurrency limits in force until
  the underlying work finishes, including after a caller times out.
- Shutdown must wait for owned work and release resources within its deadline.
  Report incomplete cleanup. Release acquired resources when initialization
  fails partway through.

### Logging

- Inject component loggers through options and default to a discard logger when
  none is supplied. Use a `component` field to identify the owner of a log.
- Prefer context-aware logging when a context is available.
- With `slog`, use `slog.Any("error", err)` for logged errors.

### Testing

- Test observable behavior and failure paths.
- Keep unit tests for `foo.go` together in `foo_test.go` in the same directory,
  even when the suite grows. Use descriptive names for package-wide, integration,
  and fuzz tests.
- Put integration and end-to-end tests in `tests/integration/` and `tests/e2e/`.
  Keep other suites in their matching directories under `tests/` when used.
- Isolate resources, clean them up, and bound waits. Prefer synchronization or
  `testing/synctest` over fixed sleeps for in-process concurrency.
- Run the checks appropriate to the changed behavior and affected callers.
  Report checks actually performed, their results, and material gaps in
  verification.

### Generated and vendored files

- Edit source definitions and generator configuration, then regenerate outputs
  through the project's workflow.
- Update vendored files through their dependency update process, preserving
  recorded versions and sources.

## Read detailed conventions on demand

Before editing or reviewing Go code, consult the sections relevant to the task
using the table below. Paths are relative to the project root. These Markdown
snapshots are installed locally by go42x; no network access is needed.

When available, use the knowledge-base tools from the `go42x` MCP server. Search
with `kwb_search`, setting `kind="documentation"`, `language="md"`, and `path_prefix`
to the selected reference. Read matching sections with `kwb_get_file` using their
source line ranges. If the tools are unavailable or the reference is not indexed,
use `rg` (ripgrep) to find headings or symbols and read the relevant source lines
with a file-reading tool. Follow cross-references when more detail is needed.
Read only the necessary sections; do not load every reference into the initial
context.

| Task or question                                 | Local reference                               | Sections to find                                                          |
| ------------------------------------------------ | --------------------------------------------- | ------------------------------------------------------------------------- |
| Choosing package names and package boundaries    | `.go42x/upstream/go/package-names.md`         | Package names, Naming package contents, Package paths, Bad package names  |
| Naming exported APIs, getters, and interfaces    | `.go42x/upstream/go/effective-go.md`          | Package names, Getters, Interface names, MixedCaps                        |
| Common review issues: names, imports, APIs       | `.go42x/upstream/go/code-review-comments.md`  | Imports, Initialisms, Interfaces, Receiver Type, Contexts, Handle Errors  |
| Choosing a Google style document                 | `.go42x/upstream/google-go/index.md`          | About, Documents                                                          |
| Readability, simplicity, formatting, consistency | `.go42x/upstream/google-go/guide.md`          | Style principles, Core guidelines, Local consistency                      |
| Detailed rules for code, comments, errors, tests | `.go42x/upstream/google-go/decisions.md`      | Naming, Commentary, Imports, Errors, Language, Contexts, Test structure   |
| Practical examples and design tradeoffs          | `.go42x/upstream/google-go/best-practices.md` | Error handling, Documentation, Function argument lists, Tests, Interfaces |
