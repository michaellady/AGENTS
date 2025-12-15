package checker

import (
	"regexp"
	"strings"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func init() {
	Register(&ClarifyingQuestions{})
}

// ClarifyingQuestions checks that agents ask clarifying questions before
// starting complex or ambiguous tasks.
// Per AGENTS.md Rule 13: "Ask the user questions before starting complex tasks"
type ClarifyingQuestions struct{}

func (c *ClarifyingQuestions) ID() string {
	return "clarifying-questions"
}

func (c *ClarifyingQuestions) Description() string {
	return "Ensures clarifying questions are asked before complex/ambiguous tasks (Rule 13)"
}

// Patterns that indicate complex or ambiguous task requests
var complexTaskPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)add\s+(user\s+)?authentication`),
	regexp.MustCompile(`(?i)implement\s+.*(feature|system|module)`),
	regexp.MustCompile(`(?i)refactor\s+`),
	regexp.MustCompile(`(?i)optimize\s+`),
	regexp.MustCompile(`(?i)improve\s+(the\s+)?performance`),
	regexp.MustCompile(`(?i)add\s+(a\s+)?new\s+(feature|endpoint|api)`),
	regexp.MustCompile(`(?i)build\s+(a\s+)?(new\s+)?\w+\s*(feature|system|module)`),
	regexp.MustCompile(`(?i)create\s+(a\s+)?(new\s+)?\w+\s*(feature|system|module|service)`),
	regexp.MustCompile(`(?i)integrate\s+`),
	regexp.MustCompile(`(?i)migrate\s+`),
	regexp.MustCompile(`(?i)redesign\s+`),
	regexp.MustCompile(`(?i)add\s+.*\s+to\s+(this|the)\s+(app|application|project)`),
}

// Patterns that indicate the agent is asking clarifying questions
var questionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)before\s+i\s+(start|begin|implement|proceed)`),
	regexp.MustCompile(`(?i)i\s+have\s+(a\s+few|some)\s+questions`),
	regexp.MustCompile(`(?i)which\s+(approach|method|option|library)`),
	regexp.MustCompile(`(?i)should\s+(i|we|it)\s+.*\?`),
	regexp.MustCompile(`(?i)would\s+you\s+(like|prefer)`),
	regexp.MustCompile(`(?i)do\s+you\s+(want|prefer|need)`),
	regexp.MustCompile(`(?i)what\s+(should|would)\s+.*\?`),
	regexp.MustCompile(`(?i)could\s+you\s+clarify`),
	regexp.MustCompile(`(?i)a\s+few\s+(clarifying\s+)?questions`),
	regexp.MustCompile(`(?i)option\s+(a|1|one).*option\s+(b|2|two)`),
	// Note: Removed overly broad `(?i)\?\s*$` pattern - it matched any question including
	// bad examples like "Should I proceed?" The other patterns are more targeted.
}

// Implementation patterns that indicate work has started without questions
var implementationPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)let\s+me\s+(start|begin|create|implement|write)`),
	regexp.MustCompile(`(?i)i('ll|'m\s+going\s+to)\s+(start|begin|create|implement|write)`),
	regexp.MustCompile(`(?i)i\s+will\s+(now\s+)?(start|begin|create|implement|write)`),
	regexp.MustCompile(`(?i)creating\s+(the|a)\s+`),
	regexp.MustCompile(`(?i)implementing\s+`),
	regexp.MustCompile(`(?i)i've\s+(created|implemented|added|written)`),
	regexp.MustCompile(`(?i)^starting\s+(the|to)\s+`),
	regexp.MustCompile(`(?i)now\s+i('ll|'m\s+going\s+to)\s+(start|begin|create|implement|write)`),
}

// Tool names that indicate implementation has started
var implementationTools = map[string]bool{
	"Write":        true,
	"Edit":         true,
	"NotebookEdit": true,
}

func (c *ClarifyingQuestions) Check(t *transcript.Transcript) []Violation {
	var violations []Violation

	// Track user messages that contain complex task requests
	type taskRequest struct {
		eventIdx int
		uuid     string
		text     string
	}
	var complexTasks []taskRequest

	// Track assistant messages
	type assistantMsg struct {
		eventIdx            int
		uuid                string
		text                string
		usesAskQuestion     bool
		asksQuestions       bool
		startsImplement     bool
		usesImplementTool   bool
	}
	var assistantMsgs []assistantMsg

	// First pass: collect user messages with complex tasks
	for i, event := range t.Events {
		userEv, ok := event.(transcript.UserEvent)
		if !ok {
			continue
		}

		var text strings.Builder
		for _, content := range userEv.Message.Content {
			if content.Type == "text" {
				text.WriteString(content.Text)
			}
		}

		msgText := text.String()
		for _, pattern := range complexTaskPatterns {
			if pattern.MatchString(msgText) {
				complexTasks = append(complexTasks, taskRequest{
					eventIdx: i,
					uuid:     userEv.UUID,
					text:     msgText,
				})
				break
			}
		}
	}

	// If no complex tasks found, nothing to check
	if len(complexTasks) == 0 {
		return violations
	}

	// Second pass: collect assistant messages and analyze them
	for i, event := range t.Events {
		assistantEv, ok := event.(transcript.AssistantEvent)
		if !ok {
			continue
		}

		var text strings.Builder
		usesAskQuestion := false
		usesImplementTool := false

		for _, content := range assistantEv.Message.Content {
			if content.Type == "text" {
				text.WriteString(content.Text)
			}
			// Check if AskUserQuestion tool is used
			if content.Type == "tool_use" && content.Name == "AskUserQuestion" {
				usesAskQuestion = true
			}
			// Check if implementation tools (Write/Edit) are used
			if content.Type == "tool_use" && implementationTools[content.Name] {
				usesImplementTool = true
			}
		}

		msgText := text.String()

		// Check if assistant asks questions
		asksQuestions := false
		for _, pattern := range questionPatterns {
			if pattern.MatchString(msgText) {
				asksQuestions = true
				break
			}
		}

		// Check if assistant starts implementing
		startsImplement := false
		for _, pattern := range implementationPatterns {
			if pattern.MatchString(msgText) {
				startsImplement = true
				break
			}
		}

		assistantMsgs = append(assistantMsgs, assistantMsg{
			eventIdx:          i,
			uuid:              assistantEv.UUID,
			text:              msgText,
			usesAskQuestion:   usesAskQuestion,
			asksQuestions:     asksQuestions,
			startsImplement:   startsImplement,
			usesImplementTool: usesImplementTool,
		})
	}

	// For each complex task, check if questions were asked before implementation
	for _, task := range complexTasks {
		// Find the first assistant response after this task
		questionAsked := false
		implementationStarted := false
		var violatingMsgUUID string

		for _, msg := range assistantMsgs {
			if msg.eventIdx <= task.eventIdx {
				continue
			}

			// Check if questions were asked (via tool or text)
			if msg.usesAskQuestion || msg.asksQuestions {
				questionAsked = true
			}

			// Check if implementation started without questions (via text or tool use)
			if (msg.startsImplement || msg.usesImplementTool) && !questionAsked {
				implementationStarted = true
				violatingMsgUUID = msg.uuid
				break
			}

			// If we find questions first, we're good
			if questionAsked {
				break
			}
		}

		if implementationStarted && !questionAsked {
			violations = append(violations, Violation{
				CheckerID: c.ID(),
				Rule:      "Rule 13",
				Severity:  SeverityWarning,
				Message:   "Started implementing complex task without asking clarifying questions",
				EventUUID: violatingMsgUUID,
				Context: map[string]string{
					"task": truncate(task.text, 100),
				},
			})
		}
	}

	return violations
}
