package checker

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func init() {
	Register(&SingleLineCommit{})
}

// SingleLineCommit checks that git commits use single-line messages.
// Per AGENTS.md: "Single-line commits only."
type SingleLineCommit struct{}

func (c *SingleLineCommit) ID() string {
	return "single-line-commit"
}

func (c *SingleLineCommit) Description() string {
	return "Ensures git commits use single-line messages (Commit Message Format)"
}

// BashInput represents the input structure for Bash tool calls.
type BashInput struct {
	Command string `json:"command"`
}

// heredocPattern matches heredoc patterns in commit commands
var heredocPattern = regexp.MustCompile(`<<['"]?EOF['"]?|<<-['"]?EOF['"]?|\$\(cat <<`)

// multilineMessagePattern matches -m with content containing newlines
var gitCommitPattern = regexp.MustCompile(`git\s+commit`)

func (c *SingleLineCommit) Check(t *transcript.Transcript) []Violation {
	var violations []Violation

	for _, tc := range t.ToolCalls {
		if tc.Name != "Bash" {
			continue
		}

		// Parse the Bash input
		var input BashInput
		if err := json.Unmarshal(tc.Input, &input); err != nil {
			continue
		}

		// Check if this is a git commit command
		if !gitCommitPattern.MatchString(input.Command) {
			continue
		}

		// Check for heredoc pattern (violation)
		if heredocPattern.MatchString(input.Command) {
			violations = append(violations, Violation{
				CheckerID:  c.ID(),
				Rule:       "Commit Message Format",
				Severity:   SeverityError,
				Message:    "Git commit uses heredoc format; use single-line -m \"message\" instead",
				EventUUID:  tc.EventUUID,
				ToolCallID: tc.ID,
				Context: map[string]string{
					"command": truncate(input.Command, 100),
				},
			})
			continue
		}

		// Check for multi-line message (newlines in the -m argument)
		if hasMultilineMessage(input.Command) {
			violations = append(violations, Violation{
				CheckerID:  c.ID(),
				Rule:       "Commit Message Format",
				Severity:   SeverityError,
				Message:    "Git commit message contains newlines; use single-line format",
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

// hasMultilineMessage checks if a git commit command has a multi-line -m message.
func hasMultilineMessage(cmd string) bool {
	// Find -m followed by quoted string
	// Look for newlines within the message
	idx := strings.Index(cmd, "-m ")
	if idx == -1 {
		idx = strings.Index(cmd, "-m\"")
		if idx == -1 {
			return false
		}
	}

	// Get the part after -m
	rest := cmd[idx+2:]
	rest = strings.TrimLeft(rest, " ")

	if len(rest) == 0 {
		return false
	}

	// Check for actual newlines in the message portion
	// Look for the quoted message
	quote := rest[0]
	if quote != '"' && quote != '\'' {
		return false
	}

	// Find the closing quote, accounting for escapes
	inEscape := false
	for i := 1; i < len(rest); i++ {
		if inEscape {
			inEscape = false
			continue
		}
		if rest[i] == '\\' {
			inEscape = true
			continue
		}
		if rest[i] == byte(quote) {
			// Found closing quote, check if message had newlines
			message := rest[1:i]
			return strings.Contains(message, "\n")
		}
	}

	return false
}

// truncate shortens a string for display.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
