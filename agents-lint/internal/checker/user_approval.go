package checker

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func init() {
	Register(&UserApproval{})
}

// UserApproval checks that user approval is requested before working on issues.
// Per AGENTS.md Rule 4: "Request approval before working on any bead issue"
type UserApproval struct{}

func (c *UserApproval) ID() string {
	return "user-approval"
}

func (c *UserApproval) Description() string {
	return "Ensures user approval is requested before working on bead issues (Rule 4)"
}

// Patterns for detecting approval-related content
var (
	// Pattern for bd update with in_progress status
	bdInProgressPattern = regexp.MustCompile(`bd\s+update\s+\S+\s+--status\s+in_progress`)

	// Patterns for approval requests in assistant messages
	approvalPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)proceed\s*\?`),
		regexp.MustCompile(`(?i)\[yes/no\]`),
		regexp.MustCompile(`(?i)ready\s+to\s+work\s+on`),
		regexp.MustCompile(`(?i)shall\s+i\s+(start|begin|proceed)`),
		regexp.MustCompile(`(?i)would\s+you\s+like\s+me\s+to`),
		regexp.MustCompile(`(?i)should\s+i\s+(start|begin|proceed)`),
	}
)

func (c *UserApproval) Check(t *transcript.Transcript) []Violation {
	var violations []Violation

	// Track assistant messages and their indices
	type msgInfo struct {
		eventIdx int
		uuid     string
		text     string
	}
	var assistantMsgs []msgInfo

	for i, event := range t.Events {
		assitantEv, ok := event.(transcript.AssistantEvent)
		if !ok {
			continue
		}

		// Collect text content from assistant
		var text strings.Builder
		for _, content := range assitantEv.Message.Content {
			if content.Type == "text" {
				text.WriteString(content.Text)
			}
		}
		if text.Len() > 0 {
			assistantMsgs = append(assistantMsgs, msgInfo{
				eventIdx: i,
				uuid:     assitantEv.UUID,
				text:     text.String(),
			})
		}
	}

	// Check each tool call for bd update --status in_progress
	for _, tc := range t.ToolCalls {
		if tc.Name != "Bash" {
			continue
		}

		var input BashInput
		if err := json.Unmarshal(tc.Input, &input); err != nil {
			continue
		}

		if !bdInProgressPattern.MatchString(input.Command) {
			continue
		}

		// Found a bd update to in_progress - check if there was an approval request
		// Look at the most recent assistant text before this tool call
		approvalFound := false

		// Find the event index for this tool call
		toolEventIdx := -1
		for i, event := range t.Events {
			assitantEv, ok := event.(transcript.AssistantEvent)
			if !ok {
				continue
			}
			for _, content := range assitantEv.Message.Content {
				if content.Type == "tool_use" && content.ID == tc.ID {
					toolEventIdx = i
					break
				}
			}
			if toolEventIdx >= 0 {
				break
			}
		}

		// Check preceding assistant messages for approval patterns
		for _, msg := range assistantMsgs {
			if msg.eventIdx >= toolEventIdx {
				continue
			}
			for _, pattern := range approvalPatterns {
				if pattern.MatchString(msg.text) {
					approvalFound = true
					break
				}
			}
			if approvalFound {
				break
			}
		}

		if !approvalFound {
			violations = append(violations, Violation{
				CheckerID:  c.ID(),
				Rule:       "Rule 4",
				Severity:   SeverityWarning,
				Message:    "Started work on bead issue without requesting user approval",
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
