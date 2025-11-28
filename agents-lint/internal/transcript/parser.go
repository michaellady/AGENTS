// Package transcript provides parsing for Claude Code stream-json transcripts.
package transcript

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// ParseFile reads and parses an NDJSON transcript file.
func ParseFile(path string) (*Transcript, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()

	var lines [][]byte
	scanner := bufio.NewScanner(f)
	// Increase buffer size for large tool outputs
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		// Make a copy since scanner reuses the buffer
		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)
		lines = append(lines, lineCopy)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read transcript: %w", err)
	}

	return parseLines(lines)
}

// ParseBytes parses NDJSON transcript data from a byte slice.
func ParseBytes(data []byte) (*Transcript, error) {
	var lines [][]byte
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			if i > start {
				lines = append(lines, data[start:i])
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}

	return parseLines(lines)
}

// parseLines parses individual NDJSON lines into a Transcript.
func parseLines(lines [][]byte) (*Transcript, error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty transcript")
	}

	t := &Transcript{
		Events: make([]any, 0, len(lines)),
	}

	// Map tool_use IDs to their results for building ToolCalls
	toolResults := make(map[string]struct {
		result  string
		isError bool
	})

	for i, line := range lines {
		// First, parse just the type field to determine event type
		var base Event
		if err := json.Unmarshal(line, &base); err != nil {
			return nil, fmt.Errorf("line %d: parse event type: %w", i+1, err)
		}

		switch base.Type {
		case "system":
			var ev SystemEvent
			if err := json.Unmarshal(line, &ev); err != nil {
				return nil, fmt.Errorf("line %d: parse system event: %w", i+1, err)
			}
			t.SessionID = ev.SessionID
			t.Model = ev.Model
			t.CWD = ev.CWD
			t.Tools = ev.Tools
			t.Events = append(t.Events, ev)

		case "assistant":
			var ev AssistantEvent
			if err := json.Unmarshal(line, &ev); err != nil {
				return nil, fmt.Errorf("line %d: parse assistant event: %w", i+1, err)
			}
			t.Events = append(t.Events, ev)

		case "user":
			var ev UserEvent
			if err := json.Unmarshal(line, &ev); err != nil {
				return nil, fmt.Errorf("line %d: parse user event: %w", i+1, err)
			}
			// Store tool results for later matching
			for _, content := range ev.Message.Content {
				if content.Type == "tool_result" && content.ToolUseID != "" {
					// Content can be string or array - extract string representation
					contentStr := extractContentString(content.Content)
					toolResults[content.ToolUseID] = struct {
						result  string
						isError bool
					}{contentStr, content.IsError}
				}
			}
			t.Events = append(t.Events, ev)

		case "result":
			var ev ResultEvent
			if err := json.Unmarshal(line, &ev); err != nil {
				return nil, fmt.Errorf("line %d: parse result event: %w", i+1, err)
			}
			t.TotalCostUSD = ev.TotalCostUSD
			t.NumTurns = ev.NumTurns
			t.IsError = ev.IsError
			t.Result = ev.Result
			t.Events = append(t.Events, ev)

		default:
			// Store unknown events as raw JSON
			var raw json.RawMessage = line
			t.Events = append(t.Events, raw)
		}
	}

	// Extract tool calls from assistant events
	t.ToolCalls = ExtractToolCalls(t, toolResults)

	return t, nil
}

// extractContentString extracts string content from raw JSON that can be a string or array.
func extractContentString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try as string first
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// Try as array of content blocks
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var result string
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				if result != "" {
					result += "\n"
				}
				result += b.Text
			}
		}
		return result
	}

	// Return as-is if neither works
	return string(raw)
}

// ExtractToolCalls extracts all tool calls from a transcript with their results.
func ExtractToolCalls(t *Transcript, toolResults map[string]struct {
	result  string
	isError bool
}) []ToolCall {
	var calls []ToolCall

	for _, event := range t.Events {
		assitantEv, ok := event.(AssistantEvent)
		if !ok {
			continue
		}

		for _, content := range assitantEv.Message.Content {
			if content.Type != "tool_use" {
				continue
			}

			tc := ToolCall{
				ID:        content.ID,
				Name:      content.Name,
				Input:     content.Input,
				EventUUID: assitantEv.UUID,
			}

			// Match with result if available
			if res, ok := toolResults[content.ID]; ok {
				tc.Result = res.result
				tc.IsError = res.isError
			}

			calls = append(calls, tc)
		}
	}

	return calls
}
