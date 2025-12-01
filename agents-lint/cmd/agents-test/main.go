// agents-test runs Claude Code test scenarios and validates behavior.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/michaellady/agents-lint/internal/runner"
	"github.com/michaellady/agents-lint/internal/scenario"

	// Register all checkers
	_ "github.com/michaellady/agents-lint/internal/checker"
)

const (
	exitOK     = 0
	exitFailed = 1
	exitError  = 2
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(exitError)
	}

	switch os.Args[1] {
	case "run":
		os.Exit(runScenarios(os.Args[2:]))
	case "list":
		os.Exit(listScenarios(os.Args[2:]))
	case "-h", "--help", "help":
		printUsage()
		os.Exit(exitOK)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(exitError)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `agents-test - Run Claude Code test scenarios

Usage:
  agents-test run [options] <scenario.yaml|scenarios/>
  agents-test list <scenarios/>

Commands:
  run     Execute scenario(s) and check results
  list    List available scenarios

Run Options:
  -agents string    Path to AGENTS.md file
  -output string    Directory to save transcripts (default ".")
  -model string     Override model (e.g., "sonnet", "haiku")
  -dry-run          Show commands without executing
  -verbose          Enable detailed output

Exit Codes:
  0  All scenarios passed
  1  One or more scenarios failed
  2  Error (invalid args, file not found, etc.)`)
}

func runScenarios(args []string) int {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	agentsPath := fs.String("agents", "", "Path to AGENTS.md file")
	outputDir := fs.String("output", ".", "Directory to save transcripts")
	model := fs.String("model", "", "Override model")
	dryRun := fs.Bool("dry-run", false, "Show commands without executing")
	verbose := fs.Bool("verbose", false, "Enable detailed output")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitError
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: scenario file or directory required")
		return exitError
	}

	path := fs.Arg(0)

	// Load scenario(s)
	var scenarios []*scenario.Scenario
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitError
	}

	if info.IsDir() {
		scenarios, err = scenario.LoadAll(path)
	} else {
		var s *scenario.Scenario
		s, err = scenario.Load(path)
		if err == nil {
			scenarios = []*scenario.Scenario{s}
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading scenarios: %v\n", err)
		return exitError
	}

	if len(scenarios) == 0 {
		fmt.Fprintln(os.Stderr, "No scenarios found")
		return exitError
	}

	// Create runner
	r := runner.New(runner.Config{
		AGENTSPath: *agentsPath,
		OutputDir:  *outputDir,
		Model:      *model,
		DryRun:     *dryRun,
		Verbose:    *verbose,
	})

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Run scenarios
	var passed, failed int
	for _, s := range scenarios {
		fmt.Printf("Running: %s\n", s.Name)
		if s.Description != "" && *verbose {
			fmt.Printf("  %s\n", s.Description)
		}

		result := r.Run(ctx, s)

		if result.Error != nil {
			fmt.Printf("  ERROR: %v\n", result.Error)
			failed++
			continue
		}

		if *dryRun {
			continue
		}

		status := "PASS"
		if !result.Passed {
			status = "FAIL"
			failed++
		} else {
			passed++
		}

		fmt.Printf("  %s (%dms, %d violations)\n", status, result.DurationMS, result.ViolationCount)
		if result.TranscriptPath != "" && *verbose {
			fmt.Printf("  Transcript: %s\n", result.TranscriptPath)
		}
	}

	if *dryRun {
		return exitOK
	}

	// Summary
	fmt.Printf("\nResults: %d passed, %d failed\n", passed, failed)

	if failed > 0 {
		return exitFailed
	}
	return exitOK
}

func listScenarios(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: scenarios directory required")
		return exitError
	}

	dir := args[0]
	scenarios, err := scenario.LoadAll(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitError
	}

	if len(scenarios) == 0 {
		fmt.Println("No scenarios found")
		return exitOK
	}

	fmt.Println("Available scenarios:")
	for _, s := range scenarios {
		absPath, _ := filepath.Abs(dir + "/" + s.Name + ".yaml")
		fmt.Printf("  %-30s %s\n", s.Name, s.Description)
		if len(s.Tags) > 0 {
			fmt.Printf("    Tags: %v\n", s.Tags)
		}
		_ = absPath
	}
	return exitOK
}
