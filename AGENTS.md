# Agent Rules & Guidelines

This document contains the rules and guidelines for Claude agents.

## Rule 1: Permission Configuration

**All Bash commands are allowed. `rm` commands require user approval.**

✅ **Allowed:** All Bash commands except `rm`
⚠️ **Requires Approval:** Any command starting with `rm`

---

## Rule 2: Task Tracking & Documentation in Beads

**ALL work tracking, documentation, architecture notes, and intermediate findings MUST be tracked in beads issue tracker.**

**CRITICAL: Beads is your documentation system. Do NOT create separate markdown files for notes, architecture, or intermediate documentation.**

### Database Initialization

**Before using beads commands, check if a beads database exists in the repository. If not, initialize one.**

**Check for database:**
```bash
# Look for .beads directory or beads.db file
ls -la .beads 2>/dev/null || ls -la beads.db 2>/dev/null
```

**Initialize if missing:**
```bash
bd init
```

**IMPORTANT:** Always run `bd init` when starting work in a new repository that doesn't have beads initialized. This creates the necessary database and configuration files.

### Verify Correct Database Location

**After initialization or when starting work, verify you're using the LOCAL repository database, not a global one.**

**Check current database location:**
```bash
# View beads stats to see which database is active
bd stats

# Check if local .beads directory exists
ls -la .beads 2>/dev/null && echo "✓ Using local database" || echo "⚠ No local database found"
```

**Verify database activity:**
```bash
# List recent issues - they should use the repository's prefix
bd list | head -5

# Check .beads directory was modified recently
ls -lth .beads/ | head -3
```

**Common issue - Using global database:**
If you're using a global beads database (e.g., from `/Users/username/dev/beads/`), issues will have a different prefix and won't be repository-specific.

**Solution:** Run `bd init` in the repository root to create a local database, then verify with `bd stats` and check for `.beads/` directory.

### Basic Commands

**Create an issue:**
```bash
bd create "Task Title" [type] [priority] -d "Description"
```

**Types:** `bug`, `feature`, `task`, `epic`, `chore`
**Priority:** `0-4` or `P0-P4` (0=critical, 4=backlog, default=2)

**Update status:**
```bash
bd update [issue-id] -s [status]
```

**Statuses:** `open`, `in_progress`, `blocked`, `closed`

**Close an issue:**
```bash
bd close [issue-id] -r "Reason for closing"
```

**Add notes:**
```bash
bd update [issue-id] --notes "Additional details"
```

**View issues:**
```bash
bd list                    # List all issues
bd list -s open            # Filter by status
bd show [issue-id]         # Show details
bd ready                   # Show ready work (no blockers)
```

### Workflow Example
```bash
# Create and start work
bd create "Implement user auth" feature 1 -d "Add JWT-based auth system"
bd update site-abc -s in_progress

# Add notes as you work
bd update site-abc --notes "Implemented login endpoint, testing logout"

# Close when complete
bd close site-abc -r "JWT auth with login/logout endpoints working"
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

### What NOT to Create
**NEVER create these types of files:**
- `NOTES.md`, `TODO.md`, `ARCHITECTURE.md`
- `IMPLEMENTATION.md`, `DESIGN.md`, `DECISIONS.md`
- `PROGRESS.md`, `WORK.md`, `PLAN.md`
- Any other intermediate documentation markdown files

**Instead:** Use `bd update [issue-id] --notes "Your documentation here"`

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
bd create "User Authentication System" epic 1 -d "MVT: Basic login/logout with session management"

# Create critical path tasks
bd create "Database schema for users table" task 1 -d "Required for MVT"
bd create "Login endpoint with session creation" task 1 -d "Core MVT functionality"
bd create "Logout endpoint" task 1 -d "Core MVT functionality"
bd create "Basic auth middleware" task 1 -d "Required to test protected routes"

# Defer non-critical items
bd create "Password reset flow" feature 2 -d "Post-MVT enhancement"
bd create "OAuth integration" feature 3 -d "Post-MVT enhancement"
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
bd create "Add user login endpoint" task 1 -d "POST /login with email/password"

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