---
id: cli-output
title: CLI output
collection: handbook
sidebar_position: 3
---

# CLI output

Commands write their results to stdout. Progress and diagnostic logs go to
stderr. `--log-level` controls logs; it never hides a command's result.

Services return results and errors. CLI adapters format results, while a shared
reporter handles JSON encoding, output failures, and errors already included in
a report. Command definitions construct and validate settings, run the service,
and pass its result to the adapter.

## Formats

Commands that support `--json` advertise it in their help. Successful JSON output
contains one feature result, without a common wrapper. Results exposed through
multiple interfaces use consistent fields.

Runtime failures without a diagnostic report produce one JSON object:

```json
{"error":"operation failed","result":null}
```

`result` contains partial operation data when available, including recovery
information. Failures already described by a diagnostic report retain that
report's format and do not print a second terminal error.

In text mode, terminal errors go to stderr. An operation may also print recovery
information on stdout. Commands that return raw content write only that content
to stdout and put continuation hints on stderr.

Help, completion, and informational output use their own text format. Protocol
servers reserve stdout for protocol messages and use stderr for logs.

## Exit status

- `0`: successful execution or diagnostic checks with only warnings.
- `1`: execution failure, failed diagnostic checks, or an output write failure.
- `2`: invalid arguments, flags, or command settings.
