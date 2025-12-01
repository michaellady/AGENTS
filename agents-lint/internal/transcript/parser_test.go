package transcript

import (
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "transcripts", "passing", "simple.ndjson")
	transcript, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify system event fields
	if transcript.SessionID != "test-session-123" {
		t.Errorf("SessionID = %q, want %q", transcript.SessionID, "test-session-123")
	}
	if transcript.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Model = %q, want %q", transcript.Model, "claude-sonnet-4-20250514")
	}
	if transcript.CWD != "/Users/test/project" {
		t.Errorf("CWD = %q, want %q", transcript.CWD, "/Users/test/project")
	}
	if len(transcript.Tools) != 3 {
		t.Errorf("len(Tools) = %d, want 3", len(transcript.Tools))
	}

	// Verify result event fields
	if transcript.NumTurns != 2 {
		t.Errorf("NumTurns = %d, want 2", transcript.NumTurns)
	}
	if transcript.IsError {
		t.Error("IsError = true, want false")
	}
	if transcript.TotalCostUSD != 0.001 {
		t.Errorf("TotalCostUSD = %f, want 0.001", transcript.TotalCostUSD)
	}

	// Verify events count
	if len(transcript.Events) != 5 {
		t.Errorf("len(Events) = %d, want 5", len(transcript.Events))
	}

	// Verify tool calls extraction
	if len(transcript.ToolCalls) != 1 {
		t.Fatalf("len(ToolCalls) = %d, want 1", len(transcript.ToolCalls))
	}

	tc := transcript.ToolCalls[0]
	if tc.Name != "Read" {
		t.Errorf("ToolCall.Name = %q, want %q", tc.Name, "Read")
	}
	if tc.ID != "tool-1" {
		t.Errorf("ToolCall.ID = %q, want %q", tc.ID, "tool-1")
	}
	if tc.Result != "package main\n\nfunc main() {}" {
		t.Errorf("ToolCall.Result = %q, want file content", tc.Result)
	}
	if tc.IsError {
		t.Error("ToolCall.IsError = true, want false")
	}
}

func TestParseBytes(t *testing.T) {
	data := []byte(`{"type":"system","subtype":"init","cwd":"/test","session_id":"s1","tools":[],"model":"test","uuid":"u1"}
{"type":"result","subtype":"success","is_error":false,"duration_ms":100,"num_turns":0,"result":"done","session_id":"s1","total_cost_usd":0.0001,"uuid":"u2"}`)

	transcript, err := ParseBytes(data)
	if err != nil {
		t.Fatalf("ParseBytes failed: %v", err)
	}

	if transcript.SessionID != "s1" {
		t.Errorf("SessionID = %q, want %q", transcript.SessionID, "s1")
	}
	if len(transcript.Events) != 2 {
		t.Errorf("len(Events) = %d, want 2", len(transcript.Events))
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("nonexistent.ndjson")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseBytes_Empty(t *testing.T) {
	_, err := ParseBytes([]byte{})
	if err == nil {
		t.Error("expected error for empty transcript")
	}
}

func TestParseBytes_InvalidJSON(t *testing.T) {
	_, err := ParseBytes([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestToolCallWithError(t *testing.T) {
	data := []byte(`{"type":"system","subtype":"init","cwd":"/test","session_id":"s1","tools":["Bash"],"model":"test","uuid":"u1"}
{"type":"assistant","message":{"id":"m1","model":"test","role":"assistant","content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"exit 1"}}],"usage":{}},"session_id":"s1","uuid":"u2"}
{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"command failed","is_error":true}]},"session_id":"s1","uuid":"u3"}
{"type":"result","subtype":"success","is_error":false,"duration_ms":100,"num_turns":1,"result":"","session_id":"s1","total_cost_usd":0.0001,"uuid":"u4"}`)

	transcript, err := ParseBytes(data)
	if err != nil {
		t.Fatalf("ParseBytes failed: %v", err)
	}

	if len(transcript.ToolCalls) != 1 {
		t.Fatalf("len(ToolCalls) = %d, want 1", len(transcript.ToolCalls))
	}

	tc := transcript.ToolCalls[0]
	if !tc.IsError {
		t.Error("ToolCall.IsError = false, want true")
	}
	if tc.Result != "command failed" {
		t.Errorf("ToolCall.Result = %q, want %q", tc.Result, "command failed")
	}
}
