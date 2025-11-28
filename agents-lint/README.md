# agents-lint

A static analysis tool for validating Claude Code transcripts against [AGENTS.md](../AGENTS.md) rules.

## Installation

```bash
cd agents-lint
go build -o agents-lint ./cmd/agents-lint
```

Or install directly:

```bash
go install github.com/michaellady/agents-lint/cmd/agents-lint@latest
```

## Quick Start

```bash
# Check a transcript
./agents-lint check transcript.ndjson

# Validate AGENTS.md structure
./agents-lint validate AGENTS.md

# List available checkers
./agents-lint list

# Run specific checkers only
./agents-lint check -checker=no-todowrite,git-branch transcript.ndjson

# JSON output for CI
./agents-lint check -format=json transcript.ndjson
```

## Usage

```
agents-lint - Validate Claude Code transcripts against AGENTS.md rules

Usage:
  agents-lint check [options] <transcript.ndjson>
  agents-lint validate [options] <AGENTS.md>
  agents-lint list [--format=json]
  agents-lint <transcript.ndjson>  (shorthand for check)

Commands:
  check      Run checkers on a transcript file
  validate   Validate AGENTS.md file structure
  list       List all available checkers

Check Options:
  -checker string   Run only specific checker(s), comma-separated
  -format string    Output format: text (default) or json
  -fail-on string   Fail on: error (default), warning, or info
  -verbose          Show detailed output (text format only)

Validate Options:
  -format string    Output format: text (default) or json

Exit Codes:
  0  All checks passed
  1  One or more violations found (at specified severity)
  2  Error (invalid args, file not found, parse error)
```

## Checkers

| Checker | Rule | Severity | Description |
|---------|------|----------|-------------|
| `no-todowrite` | Rule 2 | Error | Ensures TodoWrite tool is never used (use bd instead) |
| `single-line-commit` | Commit Format | Error | Ensures git commits use single-line messages |
| `git-branch` | Rule 3 | Error | Detects direct pushes to main/master branch |
| `context-report` | Rule 5 | Warning | Ensures context usage is reported in final response |
| `user-approval` | Rule 4 | Warning | Ensures user approval before starting work on issues |
| `commit-after-edit` | Rule 6 | Warning | Ensures file edits are followed by git commits |
| `exponential-backoff` | Rule 7 | Warning | Ensures monitoring loops use exponential backoff |
| `parallel-worktree` | Rule 8 | Warning | Ensures parallel agents use git worktrees |
| `static-types` | Rule 9 | Warning | Ensures new code uses TypeScript instead of JavaScript |

### Checker Details

#### no-todowrite
Enforces Rule 2: "Use bd for ALL task tracking. NEVER use TodoWrite."

Flags any use of the `TodoWrite` tool as an error.

#### single-line-commit
Enforces the commit message format: "Single-line commits only."

Detects:
- Heredoc patterns in git commit commands
- Multi-line messages with embedded newlines

#### git-branch
Enforces Rule 3: "NEVER commit directly to main."

Detects:
- Direct pushes to main/master (`git push origin main`)
- Force pushes to main/master (`git push -f origin main`)

#### context-report
Enforces Rule 5: "Report after every response: Context: XX% used"

Checks if the final assistant message includes context usage reporting.

#### user-approval
Enforces Rule 4: "Request approval before working on any bead issue."

Detects `bd update <ID> --status in_progress` without a preceding approval request pattern (e.g., "Proceed? [Yes/No]").

#### commit-after-edit
Enforces Rule 6: "Commit after every file change."

Tracks Edit/Write/NotebookEdit tool calls and flags if not followed by a git commit within a reasonable window (default: 15 tool calls).

## Example Output

### Text Format
```
$ ./agents-lint check transcript.ndjson
[ERROR] Rule 2: TodoWrite tool used; use bd for task tracking instead

transcript.ndjson: 1 errors, 0 warnings, 0 info
```

### JSON Format
```json
{
  "file": "transcript.ndjson",
  "checkers_run": ["commit-after-edit", "context-report", "git-branch", "no-todowrite", "single-line-commit", "user-approval"],
  "violations": [
    {
      "checker_id": "no-todowrite",
      "rule": "Rule 2",
      "severity": "error",
      "message": "TodoWrite tool used; use bd for task tracking instead",
      "event_uuid": "uuid-2",
      "tool_call_id": "tool-1"
    }
  ],
  "summary": {
    "errors": 1,
    "warnings": 0,
    "infos": 0
  }
}
```

## CI Integration

### GitHub Actions

```yaml
name: AGENTS Lint

on:
  pull_request:
    paths:
      - '**.ndjson'

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Build agents-lint
        run: |
          cd agents-lint
          go build -o agents-lint ./cmd/agents-lint

      - name: Run agents-lint
        run: |
          ./agents-lint/agents-lint check -format=json transcripts/*.ndjson
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

for file in $(git diff --cached --name-only | grep '\.ndjson$'); do
  ./agents-lint check "$file" || exit 1
done
```

## Adding Custom Checkers

1. Create a new file in `internal/checker/`:

```go
package checker

import "github.com/michaellady/agents-lint/internal/transcript"

func init() {
    Register(&MyChecker{})
}

type MyChecker struct{}

func (c *MyChecker) ID() string {
    return "my-checker"
}

func (c *MyChecker) Description() string {
    return "Checks for some condition"
}

func (c *MyChecker) Check(t *transcript.Transcript) []Violation {
    var violations []Violation

    for _, tc := range t.ToolCalls {
        // Your check logic here
        if someViolation {
            violations = append(violations, Violation{
                CheckerID:  c.ID(),
                Rule:       "Rule N",
                Severity:   SeverityError,
                Message:    "Description of violation",
                EventUUID:  tc.EventUUID,
                ToolCallID: tc.ID,
            })
        }
    }

    return violations
}
```

2. The checker auto-registers via `init()`.

3. Build and run:
```bash
go build -o agents-lint ./cmd/agents-lint
./agents-lint list  # Should show your checker
```

## Transcript Format

agents-lint expects Claude Code NDJSON transcripts with the following event types:

- `system` (subtype: `init`) - Session initialization
- `assistant` - Claude's responses with tool calls
- `user` - User messages and tool results
- `result` - Session completion

See `testdata/transcripts/` for example transcripts.

## Development

```bash
# Run tests
go test ./...

# Build
go build -o agents-lint ./cmd/agents-lint

# Test with example transcripts
./agents-lint check testdata/transcripts/passing/simple.ndjson
./agents-lint check testdata/transcripts/failing/uses-todowrite.ndjson
```

## License

MIT
