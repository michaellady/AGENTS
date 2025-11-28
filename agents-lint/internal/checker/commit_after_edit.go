package checker

import (
	"encoding/json"
	"regexp"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func init() {
	Register(&CommitAfterEdit{})
}

// CommitAfterEdit checks that file edits are followed by commits.
// Per AGENTS.md Rule 6: "Commit after every file change."
type CommitAfterEdit struct {
	// MaxToolCallsBeforeCommit is the max number of tool calls allowed between edit and commit.
	// Default is 15 if not set.
	MaxToolCallsBeforeCommit int
}

func (c *CommitAfterEdit) ID() string {
	return "commit-after-edit"
}

func (c *CommitAfterEdit) Description() string {
	return "Ensures file edits are followed by git commits (Rule 6)"
}

// editTools are tools that modify files
var editTools = map[string]bool{
	"Edit":         true,
	"Write":        true,
	"NotebookEdit": true,
}

// gitCommitPattern matches git commit commands
var gitCommitCmdPattern = regexp.MustCompile(`git\s+commit`)

func (c *CommitAfterEdit) Check(t *transcript.Transcript) []Violation {
	var violations []Violation

	maxCalls := c.MaxToolCallsBeforeCommit
	if maxCalls == 0 {
		maxCalls = 15 // Default window
	}

	// Track uncommitted edits
	var pendingEdits []pendingEdit

	for i, tc := range t.ToolCalls {
		// Check if this is an edit tool
		if editTools[tc.Name] {
			pendingEdits = append(pendingEdits, pendingEdit{
				toolCallID: tc.ID,
				eventUUID:  tc.EventUUID,
				toolName:   tc.Name,
				index:      i,
			})
			continue
		}

		// Check if this is a git commit
		if tc.Name == "Bash" {
			var input BashInput
			if err := json.Unmarshal(tc.Input, &input); err != nil {
				continue
			}

			if gitCommitCmdPattern.MatchString(input.Command) {
				// Commit found - clear pending edits
				pendingEdits = nil
				continue
			}
		}

		// Check if any pending edits are too old
		for _, edit := range pendingEdits {
			if i-edit.index > maxCalls {
				violations = append(violations, Violation{
					CheckerID:  c.ID(),
					Rule:       "Rule 6",
					Severity:   SeverityWarning,
					Message:    "File edit not followed by git commit within reasonable window",
					EventUUID:  edit.eventUUID,
					ToolCallID: edit.toolCallID,
					Context: map[string]string{
						"tool":           edit.toolName,
						"calls_since":    string(rune('0' + (i - edit.index))),
						"max_calls":      string(rune('0' + maxCalls)),
					},
				})
				// Remove this edit from pending to avoid duplicate violations
				pendingEdits = removeEdit(pendingEdits, edit.toolCallID)
			}
		}
	}

	// Check for uncommitted edits at end of transcript
	for _, edit := range pendingEdits {
		violations = append(violations, Violation{
			CheckerID:  c.ID(),
			Rule:       "Rule 6",
			Severity:   SeverityWarning,
			Message:    "File edit not committed by end of session",
			EventUUID:  edit.eventUUID,
			ToolCallID: edit.toolCallID,
			Context: map[string]string{
				"tool": edit.toolName,
			},
		})
	}

	return violations
}

// pendingEdit tracks a file edit awaiting commit.
type pendingEdit struct {
	toolCallID string
	eventUUID  string
	toolName   string
	index      int
}

// removeEdit removes an edit from the slice by tool call ID.
func removeEdit(edits []pendingEdit, id string) []pendingEdit {
	for i, e := range edits {
		if e.toolCallID == id {
			return append(edits[:i], edits[i+1:]...)
		}
	}
	return edits
}
