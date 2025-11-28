package checker

import (
	"encoding/json"
	"strings"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func init() {
	Register(&ParallelWorktree{})
}

// ParallelWorktree checks that parallel agents use git worktrees.
// Per AGENTS.md Rule 8: "Each parallel agent uses its own git worktree:
// `git worktree add ../REPO-ISSUE-ID -b ISSUE-ID main`"
type ParallelWorktree struct{}

func (c *ParallelWorktree) ID() string {
	return "parallel-worktree"
}

func (c *ParallelWorktree) Description() string {
	return "Ensures parallel agents use git worktrees (Rule 8)"
}

// exemptAgentTypes are agents that don't write code and don't need worktrees
var exemptAgentTypes = map[string]bool{
	"Explore":          true, // Read-only exploration
	"Plan":             true, // Planning only
	"claude-code-guide": true, // Documentation lookup
	"statusline-setup": true, // Configuration only
}

// isExemptAgent checks if the agent type is exempt from worktree requirement
func isExemptAgent(agentType string) bool {
	return exemptAgentTypes[agentType]
}

func (c *ParallelWorktree) Check(t *transcript.Transcript) []Violation {
	var violations []Violation

	// Track whether a worktree has been created
	worktreeCreated := false

	for _, tc := range t.ToolCalls {
		// Check for git worktree add command
		if tc.Name == "Bash" {
			var input struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal(tc.Input, &input); err != nil {
				continue
			}

			if strings.Contains(input.Command, "git worktree add") {
				worktreeCreated = true
			}
		}

		// Check for Task tool invocations
		if tc.Name == "Task" {
			var input struct {
				SubagentType string `json:"subagent_type"`
				Prompt       string `json:"prompt"`
			}
			if err := json.Unmarshal(tc.Input, &input); err != nil {
				continue
			}

			// Skip exempt agent types
			if isExemptAgent(input.SubagentType) {
				continue
			}

			// If no worktree created before this Task, flag it
			if !worktreeCreated {
				violations = append(violations, Violation{
					CheckerID:  c.ID(),
					Rule:       "Rule 8",
					Severity:   SeverityWarning,
					Message:    "Parallel agent spawned without git worktree; use `git worktree add` before spawning agents",
					EventUUID:  tc.EventUUID,
					ToolCallID: tc.ID,
				})
			}
		}
	}

	return violations
}
