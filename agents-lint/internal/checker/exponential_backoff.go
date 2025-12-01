package checker

import (
	"encoding/json"
	"regexp"
	"strconv"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func init() {
	Register(&ExponentialBackoff{})
}

// ExponentialBackoff checks that monitoring loops use exponential backoff.
// Per AGENTS.md Rule 7: "Use exponential backoff when monitoring processes
// (5s → 10s → 20s → 40s → 60s cap)."
type ExponentialBackoff struct{}

func (c *ExponentialBackoff) ID() string {
	return "exponential-backoff"
}

func (c *ExponentialBackoff) Description() string {
	return "Ensures monitoring loops use exponential backoff (Rule 7)"
}

// sleepPattern matches sleep commands with numeric seconds (anywhere in command)
var sleepPattern = regexp.MustCompile(`\bsleep\s+(\d+)`)

// monitoringCommands are commands typically used in monitoring loops
var monitoringCommands = map[string]bool{
	"kubectl get":    true,
	"kubectl describe": true,
	"docker ps":      true,
	"docker logs":    true,
	"git status":     true,
	"ps aux":         true,
	"tail -f":        true,
}

// isMonitoringCommand checks if a command looks like a monitoring/polling command
func isMonitoringCommand(cmd string) bool {
	for prefix := range monitoringCommands {
		if len(cmd) >= len(prefix) && cmd[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// isBashOutputCheck checks if this is a BashOutput tool call (checking background process)
func isBashOutputCheck(tc transcript.ToolCall) bool {
	return tc.Name == "BashOutput"
}

func (c *ExponentialBackoff) Check(t *transcript.Transcript) []Violation {
	var violations []Violation

	// Track sleep durations in sequence with their associated commands
	type sleepInfo struct {
		duration  int
		toolCall  transcript.ToolCall
		prevCmd   string // The command before this sleep
	}
	var sleeps []sleepInfo
	var lastCommand string
	var lastCommandWasPolling bool

	for _, tc := range t.ToolCalls {
		if tc.Name == "Bash" {
			var input struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal(tc.Input, &input); err != nil {
				continue
			}

			// Check for sleep command (can be standalone or embedded like "echo x && sleep 5")
			if matches := sleepPattern.FindStringSubmatch(input.Command); matches != nil {
				duration, _ := strconv.Atoi(matches[1])
				// For embedded sleep, use the full command as prevCmd
				prevCmd := lastCommand
				if prevCmd == "" {
					prevCmd = input.Command
				}
				sleeps = append(sleeps, sleepInfo{
					duration: duration,
					toolCall: tc,
					prevCmd:  prevCmd,
				})
				// Track as last command for next iteration
				lastCommand = input.Command
				continue
			}

			// Track the command
			lastCommand = input.Command
			lastCommandWasPolling = isMonitoringCommand(input.Command)
		} else if isBashOutputCheck(tc) {
			// BashOutput is also a polling pattern
			lastCommand = "BashOutput"
			lastCommandWasPolling = true
		}
	}

	// Analyze sleep pattern: need 3+ sleeps to establish a monitoring pattern
	if len(sleeps) >= 3 {
		// Check if the sleeps are between repeated commands (actual monitoring)
		// All sleeps should follow the same command pattern
		isMonitoringLoop := true
		firstPrevCmd := sleeps[0].prevCmd
		for _, s := range sleeps {
			if s.prevCmd != firstPrevCmd {
				isMonitoringLoop = false
				break
			}
		}

		// If commands vary, it's not a monitoring loop
		if !isMonitoringLoop {
			return violations
		}

		// Check if sleeps are constant (not exponential)
		isConstant := true
		for i := 1; i < len(sleeps); i++ {
			if sleeps[i].duration != sleeps[0].duration {
				isConstant = false
				break
			}
		}

		if isConstant {
			// Report violation on the third constant sleep
			violations = append(violations, Violation{
				CheckerID:  c.ID(),
				Rule:       "Rule 7",
				Severity:   SeverityWarning,
				Message:    "Monitoring loop detected with constant sleep; use exponential backoff (5s → 10s → 20s → 40s → 60s cap)",
				EventUUID:  sleeps[2].toolCall.EventUUID,
				ToolCallID: sleeps[2].toolCall.ID,
			})
		} else {
			// Check if it's proper exponential backoff
			// Allow: each sleep should be >= previous (with cap at 60)
			for i := 1; i < len(sleeps); i++ {
				prev := sleeps[i-1].duration
				curr := sleeps[i].duration

				// Allow staying at 60s cap
				if prev >= 60 && curr >= 60 {
					continue
				}

				// Current should be >= previous for backoff
				if curr < prev {
					violations = append(violations, Violation{
						CheckerID:  c.ID(),
						Rule:       "Rule 7",
						Severity:   SeverityWarning,
						Message:    "Sleep duration decreased; exponential backoff should increase (5s → 10s → 20s → 40s → 60s cap)",
						EventUUID:  sleeps[i].toolCall.EventUUID,
						ToolCallID: sleeps[i].toolCall.ID,
					})
					break
				}
			}
		}
	}

	_ = lastCommandWasPolling // May use later for enhanced detection

	return violations
}
