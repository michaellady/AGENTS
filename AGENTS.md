# Agent Rules & Guidelines

This document contains the rules and guidelines for Claude agents.

## BEFORE ANYTHING ELSE

**When starting work in a new repository, run:**
```bash
bd onboard
```

Follow the instructions provided. If the repository already has a `.beads/` directory and this document includes bd workflow information, onboarding is complete and you can skip this step.

## ⚠️ CRITICAL: Commit Message Format

**🚨 ALWAYS USE SINGLE-LINE COMMIT MESSAGES 🚨**

```bash
# ✅ CORRECT - Single line only
git commit -m "Add authentication middleware"

# ❌ WRONG - Multi-line, no heredoc needed
git commit -m "$(cat <<'EOF'
Add authentication middleware
...
EOF
)"
```

**Why?** Git hook automatically adds JIRA ID from branch name:
- You write: `"Add authentication middleware"`
- Hook transforms to: `"DAX-1234: Add authentication middleware"`

**No heredoc. No multi-line. Just single line descriptions.**

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
bd list | head -3
```

**Common issue:** Using a global beads database instead of repository-specific one. Solution: Run `bd init` in the repository root.

### Quick Reference

**Check for ready work:**
```bash
bd ready                        # Show unblocked issues
bd ready --json                 # JSON output for programmatic use
```

**Create issues:**
```bash
bd create "Issue title" -t bug|feature|task|epic|chore -p 0-4 -d "Description"
bd create "Found bug" -p 1 --deps discovered-from:PARENT-ID    # Link discovered work
```

**Issue Types:** `bug`, `feature`, `task`, `epic`, `chore`
**Priorities:** `0` (critical) → `4` (backlog), default: `2`

**Update and claim:**
```bash
bd update ISSUE-ID --status in_progress    # Claim work
bd update ISSUE-ID --priority 1            # Change priority
bd update ISSUE-ID --notes "Additional details"
```

**Statuses:** `open`, `in_progress`, `blocked`, `closed`

**Complete work:**
```bash
bd close ISSUE-ID --reason "Completed"
```

**View issues:**
```bash
bd list                  # All issues
bd list -s open          # Filter by status
bd show ISSUE-ID         # Show details
```

**Note:** Add `--json` flag to any command when you need JSON output for programmatic parsing.

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update ISSUE-ID --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" -p 1 --deps discovered-from:PARENT-ID`
5. **Complete**: `bd close ISSUE-ID --reason "Done"`
6. **Commit together**: Always commit `.beads/issues.jsonl` with code changes to keep issue state in sync

**Note:** Use `--json` flag when you need structured output for parsing (e.g., `bd ready --json`).

### Workflow Example
```bash
# Check what's ready to work on
bd ready

# Create and start work
bd create "Implement user auth" -t feature -p 1 -d "Add JWT-based auth system"
bd update AGENTS-42 --status in_progress

# Add notes as you work
bd update AGENTS-42 --notes "Implemented login endpoint, testing logout"

# Close when complete
bd close AGENTS-42 --reason "JWT auth with login/logout endpoints working"
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

**For work tracking and intermediate documentation:** Use `bd update ISSUE-ID --notes "Your documentation here"`

**Benefits:**
- ✅ Clean repository root
- ✅ Clear separation between ephemeral and permanent documentation
- ✅ Preserves planning history for archaeological research
- ✅ Reduces noise when browsing the project

### Project Documentation

**All permanent project documentation belongs in README.md - do NOT create separate documentation files.**

**README.md is the single source of truth for:**
- Project overview and purpose
- Installation and setup instructions
- Usage examples and API documentation
- Configuration options
- Contributing guidelines
- Architecture and design decisions (permanent, not ephemeral)
- Troubleshooting and FAQ

**DO NOT create these files:**
- `CONTRIBUTING.md` - add to README.md ## Contributing section
- `INSTALL.md` - add to README.md ## Installation section
- `USAGE.md` - add to README.md ## Usage section
- `API.md` - add to README.md ## API section
- `CONFIGURATION.md` - add to README.md ## Configuration section
- `TROUBLESHOOTING.md` - add to README.md ## Troubleshooting section
- `FAQ.md` - add to README.md ## FAQ section
- `ARCHITECTURE.md` (permanent) - add to README.md ## Architecture section

**Exception:** Standard repository files like `LICENSE`, `CHANGELOG.md`, `CODE_OF_CONDUCT.md`, and language-specific files (e.g., `package.json`, `go.mod`) are acceptable.

**Why consolidate in README.md?**
- ✅ Single place to find all project information
- ✅ Easier to maintain (one file vs. many)
- ✅ Better discoverability for new contributors
- ✅ Reduces documentation fragmentation
- ✅ Forces concise, well-organized documentation

**If README.md becomes too long:**
- Use clear section headers with table of contents
- Consider if the project is too complex
- Break into multiple repositories if needed
- But still keep each repo's docs in its README.md

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
bd create "User Authentication System" -t epic -p 1 -d "MVT: Basic login/logout with session management"

# Create critical path tasks
bd create "Database schema for users table" -t task -p 1 -d "Required for MVT"
bd create "Login endpoint with session creation" -t task -p 1 -d "Core MVT functionality"
bd create "Logout endpoint" -t task -p 1 -d "Core MVT functionality"
bd create "Basic auth middleware" -t task -p 1 -d "Required to test protected routes"

# Defer non-critical items
bd create "Password reset flow" -t feature -p 2 -d "Post-MVT enhancement"
bd create "OAuth integration" -t feature -p 3 -d "Post-MVT enhancement"
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
bd create "Add user login endpoint" -t task -p 1 -d "POST /login with email/password"

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
- ✅ Use `--json` flag for programmatic parsing when needed
- ✅ Link discovered work with `--deps discovered-from:PARENT-ID`
- ✅ Check `bd ready` before asking "what should I work on?"
- ✅ Store AI planning docs in `history/` directory
- ✅ Keep ALL project documentation in README.md (single source of truth)
- ✅ Commit `.beads/issues.jsonl` together with code changes
- ✅ ALWAYS use feature branch + PR workflow, even for tiny changes
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems
- ❌ Do NOT clutter repo root with planning documents
- ❌ Do NOT create separate documentation files (CONTRIBUTING.md, INSTALL.md, etc.)
- ❌ NEVER commit or push directly to main/master

---

## Rule 3: Git Branch Strategy

**CRITICAL: NEVER commit or push directly to main/master. ALL changes, no matter how small, must go through a feature branch and pull request.**

**Create a new git branch for each bead issue.**

### Branch Naming
- **Format:** `[issue-id]` (e.g., `AGENTS-42`, `site-abc`)
- Always branch from `main` unless specified otherwise
- **NEVER push directly to main/master** - all changes must go through a feature branch and PR
- This applies to ALL changes: bug fixes, typos, documentation updates, everything

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

**ALL work must go through the branch + PR workflow. No exceptions for "small changes" or "quick fixes".**

When completing work on a bead branch:
1. ✅ Commit all changes with descriptive messages
2. ✅ Push commits to the remote feature branch
3. ✅ Close the bead issue with `bd close [issue-id] -r "reason"`
4. ✅ Open a pull request with `gh pr create`
5. ✅ Monitor PR checks with `gh pr checks` and ensure they pass
6. ✅ If checks fail, fix the issues and push additional commits
7. ✅ Once all checks pass, ask the user to review the PR
8. ✅ Leave the branch for review (do NOT merge to main yourself)

**DO NOT automatically merge to main unless in greenfield mode (see below).**

**CRITICAL: After completing work on a bead branch, ALWAYS open a PR, ensure all checks pass, and ask the user to review it before merging. This applies to ALL changes, even single-line fixes.**

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

### Dependency Version Policy
**When adding new dependencies to a project, always use the latest stable version available.**

This ensures:
- Access to the newest features and improvements
- Latest security patches and bug fixes
- Reduced technical debt from outdated dependencies
- Better long-term compatibility and support

Check the package registry (npm, PyPI, crates.io, etc.) for the most recent stable release before installation.

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

---

## Landing the Plane

**When the user says "let's land the plane"**, follow this clean session-ending protocol:

1. **File beads issues for any remaining work** that needs follow-up
2. **Ensure all quality gates pass** (only if code changes were made) - run tests, linters, builds (file P0 issues if broken)
3. **Update beads issues** - close finished work, update status
4. **Sync the issue tracker carefully** - Work methodically to ensure both local and remote issues merge safely. This may require pulling, handling conflicts (sometimes accepting remote changes and re-importing), syncing the database, and verifying consistency. Be creative and patient - the goal is clean reconciliation where no issues are lost.
5. **Clean up git state** - Clear old stashes and prune dead remote branches:
   ```bash
   git stash clear                    # Remove old stashes
   git remote prune origin            # Clean up deleted remote branches
   ```
6. **Verify clean state** - Ensure all changes are committed and pushed, no untracked files remain
7. **Choose a follow-up issue for next session**
   - Provide a prompt for the user to give to you in the next session
   - Format: "Continue work on ISSUE-ID: [issue title]. [Brief context about what's been done and what's next]"

### Example "Land the Plane" Session

```bash
# 1. File remaining work
bd create "Add integration tests for sync" -t task -p 2

# 2. Run quality gates (only if code changes were made)
npm test                # or: go test, pytest, etc.
npm run lint            # or: golangci-lint run, etc.

# 3. Close finished issues
bd close AGENTS-42 --reason "Completed feature and tests passing"

# 4. Sync carefully - example workflow (adapt as needed):
git pull --rebase
# If conflicts in .beads/issues.jsonl, resolve thoughtfully:
#   - git checkout --theirs .beads/issues.jsonl (accept remote)
#   - bd import -i .beads/issues.jsonl (re-import)
#   - Or manual merge, then import
bd sync  # Export/import/verify
git push
# Repeat pull/push if needed until clean

# 5. Clean up git state
git stash clear
git remote prune origin

# 6. Verify clean state
git status

# 7. Choose next work
bd ready
bd show AGENTS-44
```

### Session Handoff Deliverables

**Then provide the user with:**

- Summary of what was completed this session
- What issues were filed for follow-up
- Status of quality gates (all passing / issues filed)
- Recommended prompt for next session

**Example handoff message:**

```
✅ Session Complete

Completed this session:
- AGENTS-42: Implemented user authentication with JWT
- AGENTS-43: Added unit tests for auth middleware

Follow-up issues filed:
- AGENTS-45: Add integration tests for auth flow (P2)
- AGENTS-46: Add password reset functionality (P3)

Quality gates: ✅ All tests passing, no lint errors

Recommended prompt for next session:
"please read and apply ../AGENTS/AGENTS.md

Continue work on AGENTS-45: Add integration tests for auth flow.
The auth middleware is complete and unit tested. Next step is to
add integration tests covering the full login/logout flow."
```
