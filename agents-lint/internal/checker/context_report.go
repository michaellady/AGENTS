package checker

import (
	"regexp"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func init() {
	Register(&ContextReport{})
}

// ContextReport checks that context usage is reported.
// Per AGENTS.md Rule 5: "Report after every response: Context: XX% used"
type ContextReport struct{}

func (c *ContextReport) ID() string {
	return "context-report"
}

func (c *ContextReport) Description() string {
	return "Ensures context usage is reported (Rule 5)"
}

// contextPattern matches "Context: XX% used" with optional token counts
var contextPattern = regexp.MustCompile(`Context:\s*\d+%\s*used`)

func (c *ContextReport) Check(t *transcript.Transcript) []Violation {
	var violations []Violation

	// Find the last assistant message with text content
	var lastTextContent string
	var lastEventUUID string

	for _, event := range t.Events {
		assitantEv, ok := event.(transcript.AssistantEvent)
		if !ok {
			continue
		}

		for _, content := range assitantEv.Message.Content {
			if content.Type == "text" && content.Text != "" {
				lastTextContent = content.Text
				lastEventUUID = assitantEv.UUID
			}
		}
	}

	// If there was a final text response, check for context report
	if lastTextContent != "" && !contextPattern.MatchString(lastTextContent) {
		violations = append(violations, Violation{
			CheckerID: c.ID(),
			Rule:      "Rule 5",
			Severity:  SeverityWarning,
			Message:   "Final response missing context usage report (Context: XX% used)",
			EventUUID: lastEventUUID,
		})
	}

	return violations
}
