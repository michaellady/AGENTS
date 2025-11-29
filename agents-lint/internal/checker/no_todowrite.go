package checker

import (
	"fmt"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func init() {
	Register(&NoTodoWrite{})
}

// NoTodoWrite checks that TodoWrite tool is never used.
// Per AGENTS.md Rule 2: "Use bd for ALL task tracking. NEVER use TodoWrite."
type NoTodoWrite struct{}

func (c *NoTodoWrite) ID() string {
	return "no-todowrite"
}

func (c *NoTodoWrite) Description() string {
	return "Ensures TodoWrite tool is never used (Rule 2: use bd instead)"
}

func (c *NoTodoWrite) Check(t *transcript.Transcript) []Violation {
	var violations []Violation

	for _, tc := range t.ToolCalls {
		if tc.Name == "TodoWrite" {
			violations = append(violations, Violation{
				CheckerID:  c.ID(),
				Rule:       "Rule 2",
				Severity:   SeverityError,
				Message:    fmt.Sprintf("TodoWrite tool used; use bd for task tracking instead"),
				EventUUID:  tc.EventUUID,
				ToolCallID: tc.ID,
			})
		}
	}

	return violations
}
