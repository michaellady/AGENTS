package checker

import (
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func TestContextReport_ID(t *testing.T) {
	c := &ContextReport{}
	if c.ID() != "context-report" {
		t.Errorf("ID() = %q, want %q", c.ID(), "context-report")
	}
}

func TestContextReport_WithReport(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"basic format", "Done!\n\n---\nContext: 15% used"},
		{"with tokens", "Done!\n\n---\nContext: 15% used (30K/200K tokens)"},
		{"inline", "Context: 50% used"},
		{"different percent", "Context: 99% used"},
	}

	c := &ContextReport{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &transcript.Transcript{
				Events: []any{
					transcript.AssistantEvent{
						Event: transcript.Event{UUID: "e1"},
						Message: transcript.AssistantMessage{
							Content: []transcript.ContentBlock{
								{Type: "text", Text: tt.text},
							},
						},
					},
				},
			}

			violations := c.Check(tr)
			if len(violations) != 0 {
				t.Errorf("expected 0 violations for %q, got %d", tt.text, len(violations))
			}
		})
	}
}

func TestContextReport_MissingReport(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "I completed the task!"},
					},
				},
			},
		},
	}

	c := &ContextReport{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.Rule != "Rule 5" {
		t.Errorf("Rule = %q, want %q", v.Rule, "Rule 5")
	}
	if v.Severity != SeverityWarning {
		t.Errorf("Severity = %v, want Warning", v.Severity)
	}
}

func TestContextReport_MultipleResponses(t *testing.T) {
	// Only the last response needs the context report
	tr := &transcript.Transcript{
		Events: []any{
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "Starting work..."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "Done!\n\n---\nContext: 25% used"},
					},
				},
			},
		},
	}

	c := &ContextReport{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations when last response has context, got %d", len(violations))
	}
}

func TestContextReport_ToolUseOnly(t *testing.T) {
	// If the last message is tool use only (no text), don't flag
	tr := &transcript.Transcript{
		Events: []any{
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "tool_use", Name: "Read"},
					},
				},
			},
		},
	}

	c := &ContextReport{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for tool-use-only message, got %d", len(violations))
	}
}

func TestContextReport_EmptyTranscript(t *testing.T) {
	tr := &transcript.Transcript{Events: []any{}}

	c := &ContextReport{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for empty transcript, got %d", len(violations))
	}
}

func TestContextReport_Registered(t *testing.T) {
	c := GetByID("context-report")
	if c == nil {
		t.Error("context-report checker not registered")
	}
}
