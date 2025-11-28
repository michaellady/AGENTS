package checker

import (
	"encoding/json"
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func TestUserApproval_ID(t *testing.T) {
	c := &UserApproval{}
	if c.ID() != "user-approval" {
		t.Errorf("ID() = %q, want %q", c.ID(), "user-approval")
	}
}

func TestUserApproval_WithApproval(t *testing.T) {
	tests := []struct {
		name        string
		approvalMsg string
	}{
		{"proceed question", "Ready to work on ISSUE-123. Proceed?"},
		{"yes/no", "Shall I start? [Yes/No]"},
		{"would you like", "Would you like me to begin?"},
		{"should I", "Should I proceed with the implementation?"},
	}

	c := &UserApproval{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bdInput, _ := json.Marshal(BashInput{Command: "bd update ISSUE-123 --status in_progress"})

			tr := &transcript.Transcript{
				Events: []any{
					transcript.AssistantEvent{
						Event: transcript.Event{UUID: "e1"},
						Message: transcript.AssistantMessage{
							Content: []transcript.ContentBlock{
								{Type: "text", Text: tt.approvalMsg},
							},
						},
					},
					transcript.AssistantEvent{
						Event: transcript.Event{UUID: "e2"},
						Message: transcript.AssistantMessage{
							Content: []transcript.ContentBlock{
								{Type: "tool_use", ID: "t1", Name: "Bash", Input: bdInput},
							},
						},
					},
				},
				ToolCalls: []transcript.ToolCall{
					{ID: "t1", Name: "Bash", Input: bdInput, EventUUID: "e2"},
				},
			}

			violations := c.Check(tr)
			if len(violations) != 0 {
				t.Errorf("expected 0 violations with approval %q, got %d", tt.approvalMsg, len(violations))
			}
		})
	}
}

func TestUserApproval_WithoutApproval(t *testing.T) {
	bdInput, _ := json.Marshal(BashInput{Command: "bd update ISSUE-123 --status in_progress"})

	tr := &transcript.Transcript{
		Events: []any{
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "I'll start working on this now."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "tool_use", ID: "t1", Name: "Bash", Input: bdInput},
					},
				},
			},
		},
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Bash", Input: bdInput, EventUUID: "e2"},
		},
	}

	c := &UserApproval{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.Rule != "Rule 4" {
		t.Errorf("Rule = %q, want %q", v.Rule, "Rule 4")
	}
	if v.Severity != SeverityWarning {
		t.Errorf("Severity = %v, want Warning", v.Severity)
	}
}

func TestUserApproval_NonBdCommands(t *testing.T) {
	// Other commands shouldn't trigger violations
	gitInput, _ := json.Marshal(BashInput{Command: "git status"})

	tr := &transcript.Transcript{
		Events: []any{
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "tool_use", ID: "t1", Name: "Bash", Input: gitInput},
					},
				},
			},
		},
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Bash", Input: gitInput, EventUUID: "e1"},
		},
	}

	c := &UserApproval{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for non-bd commands, got %d", len(violations))
	}
}

func TestUserApproval_BdOtherStatus(t *testing.T) {
	// bd update with other statuses shouldn't require approval
	bdInput, _ := json.Marshal(BashInput{Command: "bd update ISSUE-123 --status closed"})

	tr := &transcript.Transcript{
		Events: []any{
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "tool_use", ID: "t1", Name: "Bash", Input: bdInput},
					},
				},
			},
		},
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Bash", Input: bdInput, EventUUID: "e1"},
		},
	}

	c := &UserApproval{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for bd update with other status, got %d", len(violations))
	}
}

func TestUserApproval_EmptyTranscript(t *testing.T) {
	tr := &transcript.Transcript{Events: []any{}}

	c := &UserApproval{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for empty transcript, got %d", len(violations))
	}
}

func TestUserApproval_Registered(t *testing.T) {
	c := GetByID("user-approval")
	if c == nil {
		t.Error("user-approval checker not registered")
	}
}
