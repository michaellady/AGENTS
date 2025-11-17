# Agent Rules & Guidelines

This document contains the rules and guidelines for Claude agents.

## BEFORE ANYTHING ELSE

**When starting work in a new repository, run:**
```bash
bd onboard
```

Follow the instructions provided. If the repository already has a `.beads/` directory and this document includes bd workflow information, onboarding is complete and you can skip this step.

## Rule 1: Permission Configuration

**All Bash commands are allowed. `rm` commands require user approval.**

✅ **Allowed:** All Bash commands except `rm`
⚠️ **Requires Approval:** Any command starting with `rm`

---

## Rule 2: Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion
- **Your documentation system**: Use beads for architecture decisions, design notes, and intermediate findings

### Database Initialization

**If `.beads/` directory doesn't exist, initialize:**
```bash
bd init
```

**Verify you're using the LOCAL repository database:**
```bash
# Check if local .beads directory exists
ls -la .beads 2>/dev/null && echo "✓ Using local database" || echo "⚠ No local database found"

# Verify issues use the repository's prefix
bd list --no-db | head -3
```

**Common issue:** Using a global beads database instead of repository-specific one. Solution: Run `bd init` in the repository root.

### Quick Reference

**Check for ready work:**
```bash
bd ready --no-db                                    # Show unblocked issues
bd ready --json --no-db                            # JSON output for programmatic use
```

**Create issues:**
```bash
bd create "Issue title" -t bug|feature|task|epic|chore -p 0-4 -d "Description" --no-db
bd create "Found bug" -p 1 --deps discovered-from:PARENT-ID --no-db    # Link discovered work
```

**Issue Types:** `bug`, `feature`, `task`, `epic`, `chore`
**Priorities:** `0` (critical) → `4` (backlog), default: `2`

**Update and claim:**
```bash
bd update ISSUE-ID --status in_progress --no-db    # Claim work
bd update ISSUE-ID --priority 1 --no-db            # Change priority
bd update ISSUE-ID --notes "Additional details" --no-db
```

**Statuses:** `open`, `in_progress`, `blocked`, `closed`

**Complete work:**
```bash
bd close ISSUE-ID --reason "Completed" --no-db
```

**View issues:**
```bash
bd list --no-db                  # All issues
bd list -s open --no-db          # Filter by status
bd show ISSUE-ID --no-db         # Show details
```

**Note:** Use `--no-db` flag for JSONL-based operation (recommended for reliability) or `--json` flag when you need JSON output for programmatic parsing.

### Workflow for AI Agents

1. **Check ready work**: `bd ready --no-db` shows unblocked issues
2. **Claim your task**: `bd update ISSUE-ID --status in_progress --no-db`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" -p 1 --deps discovered-from:PARENT-ID --no-db`
5. **Complete**: `bd close ISSUE-ID --reason "Done" --no-db`
6. **Commit together**: Always commit `.beads/issues.jsonl` with code changes to keep issue state in sync

### Workflow Example
```bash
# Check what's ready to work on
bd ready --no-db

# Create and start work
bd create "Implement user auth" -t feature -p 1 -d "Add JWT-based auth system" --no-db
bd update AGENTS-42 --status in_progress --no-db

# Add notes as you work
bd update AGENTS-42 --notes "Implemented login endpoint, testing logout" --no-db

# Close when complete
bd close AGENTS-42 --reason "JWT auth with login/logout endpoints working" --no-db
```

### What to Track in Beads
Document EVERYTHING in beads using issue descriptions and notes:
- **Architecture decisions and design notes**
- **Research findings and technical investigations**
- **Intermediate documentation and work progress**
- Test results and fixes applied
- Build outputs and deployment info
- Bug investigations and resolutions
- Configuration changes and their impact
- Implementation details and code explanations
- Meeting notes and discussion points
- Any information you would normally put in a markdown file

### Managing AI-Generated Planning Documents

AI assistants often create planning and design documents during development. **Use a dedicated directory for these ephemeral files:**

**Recommended approach:**
- Create a `history/` directory in the project root
- Store ALL AI-generated planning/design docs in `history/`
- Keep the repository root clean and focused on permanent project files
- Only access `history/` when explicitly asked to review past planning

**Files to store in history/:**
- `PLAN.md`, `IMPLEMENTATION.md`, `ARCHITECTURE.md`
- `DESIGN.md`, `CODEBASE_SUMMARY.md`, `INTEGRATION_PLAN.md`
- `TESTING_GUIDE.md`, `TECHNICAL_DESIGN.md`, and similar files

**For work tracking and intermediate documentation:** Use `bd update ISSUE-ID --notes "Your documentation here" --no-db`

**Benefits:**
- ✅ Clean repository root
- ✅ Clear separation between ephemeral and permanent documentation
- ✅ Preserves planning history for archaeological research
- ✅ Reduces noise when browsing the project

### Auto-Sync with Git

bd automatically syncs with git:
- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)
- No manual export/import needed!
- **Always commit `.beads/issues.jsonl` together with code changes** to keep issue state in sync

### Planning Strategy

When planning work with beads (especially for epics and features):

**Focus on the critical path to a Minimum Viable Testable (MVT) project:**
- Identify the smallest set of features needed to test the core functionality
- Break down work into incremental, testable milestones
- Prioritize tasks that unblock testing and validation
- Defer nice-to-haves and optimizations until after MVT is working

**Example Planning Approach:**
```bash
# Create epic for overall feature
bd create "User Authentication System" -t epic -p 1 -d "MVT: Basic login/logout with session management" --no-db

# Create critical path tasks
bd create "Database schema for users table" -t task -p 1 -d "Required for MVT" --no-db
bd create "Login endpoint with session creation" -t task -p 1 -d "Core MVT functionality" --no-db
bd create "Logout endpoint" -t task -p 1 -d "Core MVT functionality" --no-db
bd create "Basic auth middleware" -t task -p 1 -d "Required to test protected routes" --no-db

# Defer non-critical items
bd create "Password reset flow" -t feature -p 2 -d "Post-MVT enhancement" --no-db
bd create "OAuth integration" -t feature -p 3 -d "Post-MVT enhancement" --no-db
```

**Key principle:** Get to a testable state as quickly as possible, then iterate.

### Test-Driven Development
**Write automated tests BEFORE implementation whenever possible:**
- Tests provide a clear target for what "done" looks like
- Tests give you something concrete to iterate against
- Failing tests guide implementation and catch regressions
- Tests document expected behavior

**Recommended workflow:**
```bash
# 1. Create the task
bd create "Add user login endpoint" -t task -p 1 -d "POST /login with email/password" --no-db

# 2. Write the test first (it will fail)
# Create test file with expected behavior

# 3. Run tests to confirm they fail
# This validates the test is actually testing something

# 4. Implement the feature
# Write code until tests pass

# 5. Commit when tests pass
git add . && git commit -m "Add user login endpoint with tests"
```

**Benefits:**
- Clear definition of done (tests pass)
- Confidence when refactoring
- Prevents breaking existing functionality
- Faster iteration cycles

### Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--no-db` flag for JSONL-based operation (recommended)
- ✅ Use `--json` flag for programmatic parsing when needed
- ✅ Link discovered work with `--deps discovered-from:PARENT-ID`
- ✅ Check `bd ready --no-db` before asking "what should I work on?"
- ✅ Store AI planning docs in `history/` directory
- ✅ Commit `.beads/issues.jsonl` together with code changes
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems
- ❌ Do NOT clutter repo root with planning documents

---

## Rule 3: Git Branch Strategy

**Create a new git branch for each bead issue.**

### Branch Naming
- **Format:** `[issue-id]` (e.g., `site-abc`, `roller-42`)
- Always branch from `main` unless specified otherwise
- **NEVER push directly to main** - all changes must go through a feature branch and PR

### Workflow
```bash
# Start work on issue
git checkout main
git checkout -b site-abc
bd update site-abc -s in_progress

# Make changes and commit
git add .
git commit -m "Implement feature (site-abc)"

# Push to remote feature branch
git push -u origin site-abc

# Close the bead when done
bd close site-abc -r "Feature complete and tested"
```

### Default Behavior
When completing work on a bead branch:
1. ✅ Commit all changes with descriptive messages
2. ✅ Push commits to the remote feature branch
3. ✅ Close the bead issue with `bd close [issue-id] -r "reason"`
4. ✅ Open a pull request with `gh pr create`
5. ✅ Monitor PR checks with `gh pr checks` and ensure they pass
6. ✅ If checks fail, fix the issues and push additional commits
7. ✅ Once all checks pass, ask the user to review the PR
8. ✅ Leave the branch for review (do NOT merge to main)

**DO NOT automatically merge to main unless in greenfield mode (see below).**

**IMPORTANT: After completing work on a bead branch, ALWAYS open a PR, ensure all checks pass, and ask the user to review it before merging.**

### Greenfield Mode
When working in "greenfield mode" (new projects with no review process), agents MAY merge directly to main:

```bash
# After closing the bead
bd close site-abc -r "Feature complete and tested"
git checkout main
git merge site-abc
git branch -d site-abc
```

**Only use greenfield mode auto-merge when explicitly instructed by the user.**

### Pull Request Guidelines
When creating or working with pull requests:
- Always wait for all CI/CD checks to pass before merging
- Do NOT merge PRs with failing tests, linting errors, or other check failures
- If checks fail, fix the issues and push additional commits to the feature branch
- Use `gh pr checks` to monitor the status of PR checks
- **NEVER squash commits** - preserve git history by keeping all commits intact
- Use regular merge (not squash merge or rebase merge) to maintain full commit history
- After PR is merged, delete the feature branch both locally and remotely:
  ```bash
  git branch -d [issue-id]
  git push origin --delete [issue-id]
  ```

---

## Rule 4: User Review Before Execution

**Request user approval before executing work on any bead issue.**

**Request user approval before installing any dependencies, packages, or libraries.**

### Plan Format
```
Ready to work on [issue-id]: [Title]

Problem: [description]

Plan:
- [bullet points of what will be done]

Files to modify: [list]
Estimated time: [time estimate]
Risk level: [Low/Medium/High]

Proceed? [Yes/No]
```

**IMPORTANT:** Wait for explicit user approval before proceeding. Do NOT execute without "Yes".

### Dependencies Request Format
If your implementation requires new dependencies, list them for approval:
```
Dependencies needed for [issue-id]:

- [package-name] ([version]): [reason for needing it]
- [package-name] ([version]): [reason for needing it]

Install command: [e.g., npm install, pip install, go get, brew install]

Approve installation? [Yes/No]
```

**IMPORTANT:** Do NOT install dependencies without explicit approval.

---

## Rule 5: Context Usage Reporting

**Report context usage percentage after every response.**

### Format
```
---
Context: XX% used (USED/BUDGET tokens)
```

### Example
```
---
Context: 15% used (29368/200000 tokens)
```

---

## Rule 6: Git Commit on Every Change

**Create a git commit after every file change or set of related changes in a git repository.**

### Commit Requirements
- Always commit after modifying, creating, or deleting files
- Write clear, descriptive commit messages
- Follow the repository's existing commit message style
- Include the issue ID in the commit message when working on a bead issue

### Commit Message Format
```
[Brief description of change] ([issue-id])

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

### Workflow
```bash
# After making changes
git add .
git commit -m "$(cat <<'EOF'
Add user authentication endpoint (site-abc)

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

### Push Policy
- **Main/Master branches:** Do NOT push unless explicitly requested by the user
- **Feature branches:** Automatically push commits to keep remote branch updated
- Check current branch before pushing: `git branch --show-current`

```bash
# After committing, push if on feature branch
BRANCH=$(git branch --show-current)
if [[ "$BRANCH" != "main" && "$BRANCH" != "master" ]]; then
  git push -u origin "$BRANCH"
fi
```