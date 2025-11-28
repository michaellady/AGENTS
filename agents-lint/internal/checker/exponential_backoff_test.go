package checker

import (
	"encoding/json"
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func TestExponentialBackoff_ID(t *testing.T) {
	c := &ExponentialBackoff{}
	if c.ID() != "exponential-backoff" {
		t.Errorf("ID() = %q, want %q", c.ID(), "exponential-backoff")
	}
}

func TestExponentialBackoff_Description(t *testing.T) {
	c := &ExponentialBackoff{}
	desc := c.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

// Helper to create bash command tool call
func bashCall(id, eventUUID, command string) transcript.ToolCall {
	input, _ := json.Marshal(map[string]string{"command": command})
	return transcript.ToolCall{
		ID:        id,
		Name:      "Bash",
		EventUUID: eventUUID,
		Input:     input,
	}
}

func TestExponentialBackoff_ConstantSleepViolation(t *testing.T) {
	// Rule 7: Exponential backoff required
	// Repeated constant sleep values should trigger a warning
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			bashCall("t1", "e1", "kubectl get pods"),
			bashCall("t2", "e2", "sleep 5"),
			bashCall("t3", "e3", "kubectl get pods"),
			bashCall("t4", "e4", "sleep 5"),
			bashCall("t5", "e5", "kubectl get pods"),
			bashCall("t6", "e6", "sleep 5"),
			bashCall("t7", "e7", "kubectl get pods"),
		},
	}

	c := &ExponentialBackoff{}
	violations := c.Check(tr)

	if len(violations) < 1 {
		t.Fatalf("expected at least 1 violation for constant sleep pattern, got %d", len(violations))
	}

	v := violations[0]
	if v.Rule != "Rule 7" {
		t.Errorf("Rule = %q, want %q", v.Rule, "Rule 7")
	}
	if v.Severity != SeverityWarning {
		t.Errorf("Severity = %v, want Warning", v.Severity)
	}
}

func TestExponentialBackoff_ProperBackoffAllowed(t *testing.T) {
	// Proper exponential backoff pattern: 5 → 10 → 20 → 40 → 60
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			bashCall("t1", "e1", "kubectl get pods"),
			bashCall("t2", "e2", "sleep 5"),
			bashCall("t3", "e3", "kubectl get pods"),
			bashCall("t4", "e4", "sleep 10"),
			bashCall("t5", "e5", "kubectl get pods"),
			bashCall("t6", "e6", "sleep 20"),
			bashCall("t7", "e7", "kubectl get pods"),
		},
	}

	c := &ExponentialBackoff{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for proper backoff pattern, got %d", len(violations))
	}
}

func TestExponentialBackoff_SingleSleepAllowed(t *testing.T) {
	// A single sleep without repeated pattern is not a monitoring loop
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			bashCall("t1", "e1", "npm run build"),
			bashCall("t2", "e2", "sleep 5"),
			bashCall("t3", "e3", "npm run test"),
		},
	}

	c := &ExponentialBackoff{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for single sleep, got %d", len(violations))
	}
}

func TestExponentialBackoff_TwoSleepsAllowed(t *testing.T) {
	// Two sleeps might not be a monitoring pattern
	// Require 3+ to establish pattern
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			bashCall("t1", "e1", "kubectl get pods"),
			bashCall("t2", "e2", "sleep 5"),
			bashCall("t3", "e3", "kubectl get pods"),
			bashCall("t4", "e4", "sleep 5"),
			bashCall("t5", "e5", "kubectl get pods"),
		},
	}

	c := &ExponentialBackoff{}
	violations := c.Check(tr)

	// Two identical sleeps might be tolerated, but this is borderline
	// We should flag it since it looks like polling
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for just two sleeps, got %d", len(violations))
	}
}

func TestExponentialBackoff_ThreeConstantSleepsViolation(t *testing.T) {
	// Three or more constant sleeps is clearly a monitoring pattern
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			bashCall("t1", "e1", "kubectl get pods"),
			bashCall("t2", "e2", "sleep 5"),
			bashCall("t3", "e3", "kubectl get pods"),
			bashCall("t4", "e4", "sleep 5"),
			bashCall("t5", "e5", "kubectl get pods"),
			bashCall("t6", "e6", "sleep 5"),
			bashCall("t7", "e7", "kubectl get pods"),
		},
	}

	c := &ExponentialBackoff{}
	violations := c.Check(tr)

	if len(violations) < 1 {
		t.Fatalf("expected at least 1 violation for three constant sleeps, got %d", len(violations))
	}
}

func TestExponentialBackoff_BashOutputPollingViolation(t *testing.T) {
	// Checking BashOutput repeatedly with constant delays
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Bash", EventUUID: "e1", Input: toRawJSON(map[string]any{
				"command":           "npm run build",
				"run_in_background": true,
			})},
			{ID: "t2", Name: "BashOutput", EventUUID: "e2", Input: toRawJSON(map[string]any{
				"bash_id": "shell-1",
			})},
			bashCall("t3", "e3", "sleep 5"),
			{ID: "t4", Name: "BashOutput", EventUUID: "e4", Input: toRawJSON(map[string]any{
				"bash_id": "shell-1",
			})},
			bashCall("t5", "e5", "sleep 5"),
			{ID: "t6", Name: "BashOutput", EventUUID: "e6", Input: toRawJSON(map[string]any{
				"bash_id": "shell-1",
			})},
			bashCall("t7", "e7", "sleep 5"),
			{ID: "t8", Name: "BashOutput", EventUUID: "e8", Input: toRawJSON(map[string]any{
				"bash_id": "shell-1",
			})},
		},
	}

	c := &ExponentialBackoff{}
	violations := c.Check(tr)

	if len(violations) < 1 {
		t.Fatalf("expected at least 1 violation for BashOutput polling with constant sleep, got %d", len(violations))
	}
}

func TestExponentialBackoff_VariedCommandsNotMonitoring(t *testing.T) {
	// Different commands with sleeps is not a monitoring pattern
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			bashCall("t1", "e1", "npm install"),
			bashCall("t2", "e2", "sleep 5"),
			bashCall("t3", "e3", "npm run build"),
			bashCall("t4", "e4", "sleep 5"),
			bashCall("t5", "e5", "npm run test"),
			bashCall("t6", "e6", "sleep 5"),
			bashCall("t7", "e7", "npm run lint"),
		},
	}

	c := &ExponentialBackoff{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for varied commands, got %d", len(violations))
	}
}

func TestExponentialBackoff_CapAt60Allowed(t *testing.T) {
	// Backoff capped at 60s is allowed: 5 → 10 → 20 → 40 → 60 → 60
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			bashCall("t1", "e1", "kubectl get pods"),
			bashCall("t2", "e2", "sleep 5"),
			bashCall("t3", "e3", "kubectl get pods"),
			bashCall("t4", "e4", "sleep 10"),
			bashCall("t5", "e5", "kubectl get pods"),
			bashCall("t6", "e6", "sleep 20"),
			bashCall("t7", "e7", "kubectl get pods"),
			bashCall("t8", "e8", "sleep 40"),
			bashCall("t9", "e9", "kubectl get pods"),
			bashCall("t10", "e10", "sleep 60"),
			bashCall("t11", "e11", "kubectl get pods"),
			bashCall("t12", "e12", "sleep 60"),
			bashCall("t13", "e13", "kubectl get pods"),
		},
	}

	c := &ExponentialBackoff{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for capped backoff at 60s, got %d", len(violations))
	}
}

func TestExponentialBackoff_Registered(t *testing.T) {
	c := GetByID("exponential-backoff")
	if c == nil {
		t.Error("exponential-backoff checker not registered")
	}
}
