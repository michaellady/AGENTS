// agents-lint validates Claude Code transcripts against AGENTS.md rules.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: agents-lint <transcript.ndjson>")
		os.Exit(1)
	}

	// TODO: Implement transcript parsing and checker execution
	fmt.Println("agents-lint: not yet implemented")
}
