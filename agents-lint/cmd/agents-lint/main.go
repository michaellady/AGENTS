// agents-lint validates Claude Code transcripts against AGENTS.md rules.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/michaellady/agents-lint/internal/checker"
	"github.com/michaellady/agents-lint/internal/report"
	"github.com/michaellady/agents-lint/internal/transcript"

	// Register all checkers
	_ "github.com/michaellady/agents-lint/internal/checker"
)

const (
	exitOK         = 0
	exitViolations = 1
	exitError      = 2
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(exitError)
	}

	switch os.Args[1] {
	case "check":
		os.Exit(runCheck(os.Args[2:]))
	case "list":
		os.Exit(runList(os.Args[2:]))
	case "-h", "--help", "help":
		printUsage()
		os.Exit(exitOK)
	default:
		// Assume it's a file path for backwards compatibility
		os.Exit(runCheck(os.Args[1:]))
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `agents-lint - Validate Claude Code transcripts against AGENTS.md rules

Usage:
  agents-lint check [options] <transcript.ndjson>
  agents-lint list [--format=json]
  agents-lint <transcript.ndjson>  (shorthand for check)

Commands:
  check    Run checkers on a transcript file
  list     List all available checkers

Check Options:
  -checker string   Run only specific checker(s), comma-separated
  -format string    Output format: text (default) or json
  -fail-on string   Fail on: error (default), warning, or info
  -verbose          Show detailed output (text format only)

Exit Codes:
  0  All checks passed
  1  One or more violations found (at specified severity)
  2  Error (invalid args, file not found, parse error)`)
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	checkerFlag := fs.String("checker", "", "Run only specific checker(s), comma-separated")
	format := fs.String("format", "text", "Output format: text or json")
	failOn := fs.String("fail-on", "error", "Fail on: error, warning, or info")
	verbose := fs.Bool("verbose", false, "Show detailed output (text format only)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitError
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: transcript file path required")
		return exitError
	}

	path := fs.Arg(0)
	t, err := transcript.ParseFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing transcript: %v\n", err)
		return exitError
	}

	var result *checker.Result
	if *checkerFlag != "" {
		ids := strings.Split(*checkerFlag, ",")
		result = checker.RunByIDs(t, ids)
	} else {
		result = checker.RunAll(t)
	}
	result.TranscriptPath = path

	// Output results
	switch *format {
	case "json":
		if err := report.WriteJSON(os.Stdout, result); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON: %v\n", err)
			return exitError
		}
	case "text":
		report.WriteText(os.Stdout, result, *verbose)
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *format)
		return exitError
	}

	// Determine exit code based on --fail-on
	errors, warnings, infos := result.Summary()
	switch *failOn {
	case "error":
		if errors > 0 {
			return exitViolations
		}
	case "warning":
		if errors > 0 || warnings > 0 {
			return exitViolations
		}
	case "info":
		if errors > 0 || warnings > 0 || infos > 0 {
			return exitViolations
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown fail-on value: %s\n", *failOn)
		return exitError
	}

	return exitOK
}

func runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	format := fs.String("format", "text", "Output format: text or json")
	_ = fs.Parse(args)

	checkers := checker.GetAll()

	if *format == "json" {
		type checkerInfo struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		}
		list := make([]checkerInfo, len(checkers))
		for i, c := range checkers {
			list[i] = checkerInfo{ID: c.ID(), Description: c.Description()}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(list)
		return exitOK
	}

	if len(checkers) == 0 {
		fmt.Println("No checkers registered")
		return exitOK
	}

	fmt.Println("Available checkers:")
	for _, c := range checkers {
		fmt.Printf("  %-20s %s\n", c.ID(), c.Description())
	}
	return exitOK
}
