# Agent Rules & Guidelines

## BEFORE ANYTHING ELSE
Run `bd onboard` when starting work in a new repository. Skip if `.beads/` exists.

## Commit Message Format
**Single-line commits only.** Git hook adds JIRA ID from branch name automatically.
```bash
git commit -m "Add authentication middleware"  # ✅ Correct
```

## Rule 1: Permissions
All Bash commands allowed. `rm` requires user approval.

## Rule 2: Issue Tracking with bd
**Use bd for ALL task tracking. NEVER use TodoWrite or TODO comments.**

```bash
bd ready                              # Show unblocked work
bd create "Title" -t task -p 1 -d "Description"  # Create issue
bd update ID --status in_progress     # Claim work
bd close ID --reason "Done"           # Complete work
```

**Types:** `bug`, `feature`, `task`, `epic`, `chore`
**Priorities:** `0` (critical) → `4` (backlog)
**Statuses:** `open`, `in_progress`, `blocked`, `closed`

Always commit `.beads/issues.jsonl` with code changes.

## Rule 3: Git Branch Strategy
**NEVER commit directly to main. ALL changes go through feature branch + PR.**

```bash
git checkout -b ISSUE-ID              # Create branch from main
# ... make changes ...
git push -u origin ISSUE-ID           # Push to remote
gh pr create                          # Open PR
# Wait for checks to pass, then ask user to review
```

After PR merged, delete branch: `git branch -d ISSUE-ID && git push origin --delete ISSUE-ID`

## Rule 4: User Review Before Execution
**Request approval before working on any bead issue or installing dependencies.**

```
Ready to work on [ID]: [Title]
Plan: [bullet points]
Files to modify: [list]
Proceed? [Yes/No]
```

## Rule 5: Context Usage Reporting
Report after every response:
```
---
Context: XX% used (USED/BUDGET tokens)
```

## Rule 6: Git Commit on Every Change
Commit after every file change. Include `.beads/issues.jsonl`.
```bash
git commit -m "Brief description (ISSUE-ID)"
```
Push automatically on feature branches, never on main without approval.

## Rule 7: Monitoring
Use exponential backoff when monitoring processes (5s → 10s → 20s → 40s → 60s cap).
Run monitors as background processes when possible.

## Rule 8: Parallel Work
Each parallel agent uses its own git worktree: `git worktree add ../REPO-ISSUE-ID -b ISSUE-ID main`

## Rule 9: Prefer Static Types
New projects: Use Go, Kotlin, TypeScript, or Rust.
Scripting languages: Always use type hints (Python) or TypeScript (not JS).

## Rule 10: Code Search with auggie-mcp
**Use `mcp__auggie-mcp__codebase-retrieval` as PRIMARY tool for code understanding.**

- Semantic search: "Where is authentication handled?"
- Before editing: Query all related symbols in one call
- Understanding architecture: "How does X connect to Y?"

**Use Grep/Glob instead for:** exact string matching, finding all references to a known identifier, file pattern matching.

## Rule 11: Protect Secrets
**Never print secret values unless explicitly asked.**

Secrets include: API keys, tokens, passwords, private keys, credentials, connection strings with passwords.

When encountering secrets:
- Confirm the secret exists without revealing the value
- Use placeholders: `API_KEY=<redacted>` or `password=***`
- If user explicitly asks to see the value, comply with a warning

## Rule 12: JIRA-Named Repository Handling
**When working in a JIRA-named repository (like "AGENTS"), do NOT attempt to create new git repositories.**

Instead:
- Work only within existing repositories found in the workspace
- Commit changes to the appropriate sub-repositories within the workspace
- Respect the existing git structure and repository boundaries
- Use `git status` and `git remote -v` to identify the correct repository context before making commits

## Landing the Plane
When user says "land the plane":
1. File beads for remaining work
2. Run quality gates (tests, lint) if code changed
3. Close finished issues
4. Commit and push beads changes
5. Clean up: `git stash clear && git remote prune origin`
6. Provide session summary and recommended next prompt (first line must be: `please read and apply /Users/mikelady/dev/AGENTS/AGENTS.md`)

## Pass the Baton
When user says "pass the baton": Execute "Land the Plane", then spawn a new agent with continuation prompt using the Task tool.

---

**Full documentation with examples:** See [AGENTS-REFERENCE.md](./AGENTS-REFERENCE.md)
