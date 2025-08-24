# Project Conventions

## Overview

**IMPORTANT:** These conventions are mandatory unless explicitly overridden.

### Core Rules

1. **[IGNORE] blocks** - Skip any content between `[IGNORE]` and `[/IGNORE]` tags
2. **Reference existing code** - Always examine similar files before creating new ones
3. **Follow /CONVENTIONS.md** - This file contains project-specific standards that must be followed

### When Deviating from Conventions

If you need to violate a convention:

1. **STOP** and explain why the deviation is necessary
2. **ASK** for explicit approval before proceeding
3. **DOCUMENT** the approved change in `/CONVENTIONS.md` after implementation

### Priority Order

1. Project-specific conventions in `/CONVENTIONS.md`
2. Language idioms and best practices
3. Team preferences (when explicitly stated)
4. General clean code principles

# About

This document is a repository of conventions and rules used by this project.

## Foundation

* https://google.github.io/eng-practices/
* https://google.github.io/styleguide/go/decisions.html
* https://sre.google/sre-book/table-of-contents/
* https://www.conventionalcommits.org/en/v1.0.0/
* https://semver.org/

## Review

## Project Management

* tooling versions
* release process

## SVC

* branch naming
* commit message
* pull request names and description
* tag naming
* sub-module tags
* always prefer merge commits to rebase (disable rebase)
* .gitignore -> current dir / .gitkeep

## Golang

* upgrading go version
* import order
* panic recovery
* observability (tracing,tracing,metrics)
* protocol
* api versioning
* //go:generate mockgen -> always local binary
* v for validation tag
* db for db column name tag
* pass logger is dependancy injection with component field, but can be used globally where needed
* WithTransaction should NOT be used in repository level
* use `slog.Any("error", err)` for slog errors
* log.fatal can be used only during init phase in main functions
* logger should be passed as option, if not passed, must default to noop logger
* string == "" vs len(string) == 0
* log fields with dash, metric labels with underscore
* always use xContext() version of slog methods where context is available
* github.com/hasansino/go42/internal/tools should never import anything from internal
* retry pattern
* naming interfaces and generating mocks
* use `any` instead of `interface{}` in function signatures
* `context.Context` -> ctx but `echo.Context` -> c
* put technical phrases in backticks in comments to avoid linting issues
* `fmt.Errorf` vs `errors.Wrap` (collides vs std errors)
* use power of 2 for buffer sizing, implemented using bitwise shift operator
* using golines
* never use anonymous interfaces
* never use casting to anonymous interfaces
* never define types inside functions
* never use anonymous structs

## Miscellaneous

* yaml vs yml
* migration file naming
* using @see @todo @fixme @note etc. in comments
* tools configuration files should be in etc directory
* migrations should be idempotent
* sql lowercase -> it is a choice
* always leave empty lines at the end of files
* usage of `// ---``
* never expose IDs -> expose UUIDs
* always leave trailing newline for text files
