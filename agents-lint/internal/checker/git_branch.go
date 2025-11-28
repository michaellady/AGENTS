package checker

import (
	"encoding/json"
	"regexp"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func init() {
	Register(&GitBranch{})
}

// GitBranch checks git branch workflow compliance.
// Per AGENTS.md Rule 3: "NEVER commit directly to main. ALL changes go through feature branch + PR."
type GitBranch struct{}

func (c *GitBranch) ID() string {
	return "git-branch"
}

func (c *GitBranch) Description() string {
	return "Ensures proper git branch workflow (Rule 3: no direct commits to main)"
}

// Patterns for detecting branch violations
var (
	// Direct push to main/master
	pushMainPattern = regexp.MustCompile(`git\s+push\s+(-[^\s]+\s+)*origin\s+(main|master)\b`)
	// Push with -f to main/master
	forcePushMainPattern = regexp.MustCompile(`git\s+push\s+.*(-f|--force).*origin\s+(main|master)\b`)
	// Checkout main/master with intent to commit
	checkoutMainPattern = regexp.MustCompile(`git\s+checkout\s+(main|master)\s*$`)
)

func (c *GitBranch) Check(t *transcript.Transcript) []Violation {
	var violations []Violation

	for _, tc := range t.ToolCalls {
		if tc.Name != "Bash" {
			continue
		}

		var input BashInput
		if err := json.Unmarshal(tc.Input, &input); err != nil {
			continue
		}

		// Check for force push to main (most severe)
		if forcePushMainPattern.MatchString(input.Command) {
			violations = append(violations, Violation{
				CheckerID:  c.ID(),
				Rule:       "Rule 3",
				Severity:   SeverityError,
				Message:    "Force push to main/master branch detected; this is extremely dangerous",
				EventUUID:  tc.EventUUID,
				ToolCallID: tc.ID,
				Context: map[string]string{
					"command": truncate(input.Command, 100),
				},
			})
			continue
		}

		// Check for direct push to main
		if pushMainPattern.MatchString(input.Command) {
			violations = append(violations, Violation{
				CheckerID:  c.ID(),
				Rule:       "Rule 3",
				Severity:   SeverityError,
				Message:    "Direct push to main/master branch; use feature branch + PR instead",
				EventUUID:  tc.EventUUID,
				ToolCallID: tc.ID,
				Context: map[string]string{
					"command": truncate(input.Command, 100),
				},
			})
		}
	}

	return violations
}
