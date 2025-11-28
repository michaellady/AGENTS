package checker

import (
	"encoding/json"
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func TestParallelWorktree_ID(t *testing.T) {
	c := &ParallelWorktree{}
	if c.ID() != "parallel-worktree" {
		t.Errorf("ID() = %q, want %q", c.ID(), "parallel-worktree")
	}
}

func TestParallelWorktree_Description(t *testing.T) {
	c := &ParallelWorktree{}
	desc := c.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

// Helper to create Task tool call
func taskCall(id, eventUUID, prompt string) transcript.ToolCall {
	input, _ := json.Marshal(map[string]string{
		"prompt":        prompt,
		"subagent_type": "general-purpose",
	})
	return transcript.ToolCall{
		ID:        id,
		Name:      "Task",
		EventUUID: eventUUID,
		Input:     input,
	}
}

func TestParallelWorktree_TaskWithoutWorktreeViolation(t *testing.T) {
	// Rule 8: Each parallel agent should use its own git worktree
	// Spawning a Task without prior worktree setup is a violation
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			taskCall("t1", "e1", "Implement feature X"),
		},
	}

	c := &ParallelWorktree{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for Task without worktree, got %d", len(violations))
	}

	v := violations[0]
	if v.Rule != "Rule 8" {
		t.Errorf("Rule = %q, want %q", v.Rule, "Rule 8")
	}
	if v.Severity != SeverityWarning {
		t.Errorf("Severity = %v, want Warning", v.Severity)
	}
	if v.ToolCallID != "t1" {
		t.Errorf("ToolCallID = %q, want %q", v.ToolCallID, "t1")
	}
}

func TestParallelWorktree_TaskWithWorktreeAllowed(t *testing.T) {
	// Proper pattern: create worktree before spawning Task
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			bashCall("t1", "e1", "git worktree add ../myrepo-ISSUE-123 -b ISSUE-123 main"),
			taskCall("t2", "e2", "Implement feature X in the worktree"),
		},
	}

	c := &ParallelWorktree{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations when worktree is set up, got %d", len(violations))
	}
}

func TestParallelWorktree_MultipleTasksOneWorktree(t *testing.T) {
	// One worktree covers all subsequent Tasks (within same branch context)
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			bashCall("t1", "e1", "git worktree add ../myrepo-ISSUE-123 -b ISSUE-123 main"),
			taskCall("t2", "e2", "First parallel task"),
			taskCall("t3", "e3", "Second parallel task"),
		},
	}

	c := &ParallelWorktree{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for multiple tasks after worktree, got %d", len(violations))
	}
}

func TestParallelWorktree_ExploreAgentExempt(t *testing.T) {
	// Explore agents are for reading code, don't need worktrees
	input, _ := json.Marshal(map[string]string{
		"prompt":        "Find all API endpoints",
		"subagent_type": "Explore",
	})
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "t1",
				Name:      "Task",
				EventUUID: "e1",
				Input:     input,
			},
		},
	}

	c := &ParallelWorktree{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for Explore agent, got %d", len(violations))
	}
}

func TestParallelWorktree_PlanAgentExempt(t *testing.T) {
	// Plan agents are for planning, don't need worktrees
	input, _ := json.Marshal(map[string]string{
		"prompt":        "Plan the implementation",
		"subagent_type": "Plan",
	})
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "t1",
				Name:      "Task",
				EventUUID: "e1",
				Input:     input,
			},
		},
	}

	c := &ParallelWorktree{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for Plan agent, got %d", len(violations))
	}
}

func TestParallelWorktree_ClaudeCodeGuideExempt(t *testing.T) {
	// claude-code-guide agents are for documentation lookup, don't need worktrees
	input, _ := json.Marshal(map[string]string{
		"prompt":        "How do I use hooks?",
		"subagent_type": "claude-code-guide",
	})
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "t1",
				Name:      "Task",
				EventUUID: "e1",
				Input:     input,
			},
		},
	}

	c := &ParallelWorktree{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for claude-code-guide agent, got %d", len(violations))
	}
}

func TestParallelWorktree_GeneralPurposeRequiresWorktree(t *testing.T) {
	// general-purpose agents can write code, need worktrees
	input, _ := json.Marshal(map[string]string{
		"prompt":        "Implement the authentication system",
		"subagent_type": "general-purpose",
	})
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "t1",
				Name:      "Task",
				EventUUID: "e1",
				Input:     input,
			},
		},
	}

	c := &ParallelWorktree{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for general-purpose agent without worktree, got %d", len(violations))
	}
}

func TestParallelWorktree_NoTasksNoViolation(t *testing.T) {
	// No Task tool calls means no violations
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			bashCall("t1", "e1", "npm run build"),
			{ID: "t2", Name: "Read", EventUUID: "e2"},
			{ID: "t3", Name: "Write", EventUUID: "e3"},
		},
	}

	c := &ParallelWorktree{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations when no Task calls, got %d", len(violations))
	}
}

func TestParallelWorktree_WorktreeVariations(t *testing.T) {
	// Different valid worktree command formats
	tests := []struct {
		name    string
		command string
	}{
		{"basic", "git worktree add ../myrepo-ISSUE-1 -b ISSUE-1 main"},
		{"with path spaces", "git worktree add \"../my repo-ISSUE-1\" -b ISSUE-1 main"},
		{"different base", "git worktree add ../feature -b feature-branch develop"},
		{"short form", "git worktree add ../work"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &transcript.Transcript{
				ToolCalls: []transcript.ToolCall{
					bashCall("t1", "e1", tt.command),
					taskCall("t2", "e2", "Do work"),
				},
			}

			c := &ParallelWorktree{}
			violations := c.Check(tr)

			if len(violations) != 0 {
				t.Errorf("expected 0 violations for worktree command %q, got %d", tt.command, len(violations))
			}
		})
	}
}

func TestParallelWorktree_Registered(t *testing.T) {
	c := GetByID("parallel-worktree")
	if c == nil {
		t.Error("parallel-worktree checker not registered")
	}
}
