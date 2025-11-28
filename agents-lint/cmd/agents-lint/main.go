// agents-lint validates Claude Code transcripts against AGENTS.md rules.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/michaellady/agents-lint/internal/checker"
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
		os.Exit(runList())
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
  agents-lint list
  agents-lint <transcript.ndjson>  (shorthand for check)

Commands:
  check    Run checkers on a transcript file
  list     List all available checkers

Check Options:
  -checker string   Run only specific checker(s), comma-separated
  -verbose          Show detailed output

Exit Codes:
  0  All checks passed
  1  One or more violations found
  2  Error (invalid args, file not found, parse error)`)
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	checkerFlag := fs.String("checker", "", "Run only specific checker(s), comma-separated")
	verbose := fs.Bool("verbose", false, "Show detailed output")

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
	errors, warnings, infos := result.Summary()

	if *verbose || len(result.Violations) > 0 {
		for _, v := range result.Violations {
			severity := strings.ToUpper(v.Severity.String())
			fmt.Printf("[%s] %s: %s\n", severity, v.Rule, v.Message)
			if v.ToolCallID != "" && *verbose {
				fmt.Printf("  Tool call: %s\n", v.ToolCallID)
			}
		}
	}

	if *verbose {
		fmt.Printf("\nCheckers run: %s\n", strings.Join(result.CheckersRun, ", "))
	}

	fmt.Printf("\n%s: %d errors, %d warnings, %d info\n", path, errors, warnings, infos)

	if result.HasErrors() {
		return exitViolations
	}
	return exitOK
}

func runList() int {
	checkers := checker.GetAll()

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
