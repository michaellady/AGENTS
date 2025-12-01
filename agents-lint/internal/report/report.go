// Package report provides output formatting for checker results.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/michaellady/agents-lint/internal/checker"
)

// JSONReport is the structured output format.
type JSONReport struct {
	File        string          `json:"file"`
	CheckersRun []string        `json:"checkers_run"`
	Violations  []JSONViolation `json:"violations"`
	Summary     Summary         `json:"summary"`
}

// JSONViolation is a violation in JSON format.
type JSONViolation struct {
	CheckerID  string            `json:"checker_id"`
	Rule       string            `json:"rule"`
	Severity   string            `json:"severity"`
	Message    string            `json:"message"`
	EventUUID  string            `json:"event_uuid,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
}

// Summary contains violation counts.
type Summary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
}

// WriteJSON outputs the result as JSON.
func WriteJSON(w io.Writer, result *checker.Result) error {
	errors, warnings, infos := result.Summary()

	report := JSONReport{
		File:        result.TranscriptPath,
		CheckersRun: result.CheckersRun,
		Violations:  make([]JSONViolation, len(result.Violations)),
		Summary: Summary{
			Errors:   errors,
			Warnings: warnings,
			Infos:    infos,
		},
	}

	for i, v := range result.Violations {
		report.Violations[i] = JSONViolation{
			CheckerID:  v.CheckerID,
			Rule:       v.Rule,
			Severity:   v.Severity.String(),
			Message:    v.Message,
			EventUUID:  v.EventUUID,
			ToolCallID: v.ToolCallID,
			Context:    v.Context,
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// WriteText outputs the result as human-readable text.
func WriteText(w io.Writer, result *checker.Result, verbose bool) {
	errors, warnings, infos := result.Summary()

	// Print violations
	if verbose || len(result.Violations) > 0 {
		for _, v := range result.Violations {
			severity := strings.ToUpper(v.Severity.String())
			fmt.Fprintf(w, "[%s] %s: %s\n", severity, v.Rule, v.Message)
			if v.ToolCallID != "" && verbose {
				fmt.Fprintf(w, "  Tool call: %s\n", v.ToolCallID)
			}
			if len(v.Context) > 0 && verbose {
				for k, val := range v.Context {
					fmt.Fprintf(w, "  %s: %s\n", k, val)
				}
			}
		}
	}

	if verbose {
		fmt.Fprintf(w, "\nCheckers run: %s\n", strings.Join(result.CheckersRun, ", "))
	}

	fmt.Fprintf(w, "\n%s: %d errors, %d warnings, %d info\n", result.TranscriptPath, errors, warnings, infos)
}
