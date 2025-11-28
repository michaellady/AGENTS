// Package transcript defines types for parsing Claude Code stream-json transcripts.
package transcript

import "encoding/json"

// Event represents a single NDJSON line from stream-json output.
// The Type field determines which specific event struct to unmarshal into.
type Event struct {
	Type      string `json:"type"`                 // "system", "assistant", "user", "result"
	Subtype   string `json:"subtype,omitempty"`    // "init" for system, "success"/"error" for result
	SessionID string `json:"session_id,omitempty"` // Unique session identifier
	UUID      string `json:"uuid,omitempty"`       // Unique event identifier
}

// SystemEvent is emitted at the start of a session with configuration info.
type SystemEvent struct {
	Event
	CWD              string   `json:"cwd"`
	Tools            []string `json:"tools"`
	MCPServers       []string `json:"mcp_servers"`
	Model            string   `json:"model"`
	PermissionMode   string   `json:"permissionMode"`
	SlashCommands    []string `json:"slash_commands"`
	ClaudeCodeVersion string  `json:"claude_code_version"`
}

// AssistantEvent contains a message from the assistant (Claude).
type AssistantEvent struct {
	Event
	Message          AssistantMessage `json:"message"`
	ParentToolUseID  *string          `json:"parent_tool_use_id"`
}

// AssistantMessage is the core message structure from the assistant.
type AssistantMessage struct {
	ID         string         `json:"id"`
	Model      string         `json:"model"`
	Role       string         `json:"role"`
	Content    []ContentBlock `json:"content"`
	StopReason *string        `json:"stop_reason"`
	Usage      Usage          `json:"usage"`
}

// ContentBlock represents either text or tool_use content.
type ContentBlock struct {
	Type  string          `json:"type"` // "text" or "tool_use"
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`    // Tool use ID
	Name  string          `json:"name,omitempty"`  // Tool name
	Input json.RawMessage `json:"input,omitempty"` // Tool input as raw JSON
}

// Usage contains token usage information.
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// UserEvent contains a message from the user (typically tool results).
type UserEvent struct {
	Event
	Message         UserMessage `json:"message"`
	ParentToolUseID *string     `json:"parent_tool_use_id"`
	ToolUseResult   string      `json:"tool_use_result,omitempty"` // Summary of tool result
}

// UserMessage is the message structure from the user.
type UserMessage struct {
	Role    string             `json:"role"`
	Content []UserContentBlock `json:"content"`
}

// UserContentBlock represents user content, typically tool results.
type UserContentBlock struct {
	Type      string `json:"type"`                  // "tool_result" or "text"
	ToolUseID string `json:"tool_use_id,omitempty"` // References the tool_use block
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// ResultEvent is emitted at the end of a session with summary info.
type ResultEvent struct {
	Event
	IsError           bool              `json:"is_error"`
	DurationMS        int               `json:"duration_ms"`
	DurationAPIMS     int               `json:"duration_api_ms"`
	NumTurns          int               `json:"num_turns"`
	Result            string            `json:"result"`
	TotalCostUSD      float64           `json:"total_cost_usd"`
	Usage             Usage             `json:"usage"`
	ModelUsage        map[string]Usage  `json:"modelUsage"`
	PermissionDenials []PermissionDenial `json:"permission_denials,omitempty"`
}

// PermissionDenial records when a tool was blocked.
type PermissionDenial struct {
	ToolName  string          `json:"tool_name"`
	ToolUseID string          `json:"tool_use_id"`
	ToolInput json.RawMessage `json:"tool_input"`
}

// ToolCall represents a parsed tool invocation for checker analysis.
type ToolCall struct {
	ID        string          // Tool use ID
	Name      string          // Tool name (e.g., "Bash", "Read", "Write")
	Input     json.RawMessage // Raw JSON input
	Result    string          // Tool result content
	IsError   bool            // Whether the tool returned an error
	EventUUID string          // UUID of the assistant event containing this call
}

// Transcript represents a complete parsed session.
type Transcript struct {
	SessionID    string
	Model        string
	CWD          string
	Tools        []string
	Events       []any       // All events in order
	ToolCalls    []ToolCall  // Extracted tool calls for easy iteration
	TotalCostUSD float64
	NumTurns     int
	IsError      bool
	Result       string
}
