Review all changes on the current branch for quality, consistency, and spec compliance.

## Step 1: Gather context

1. Identify the current branch: `git branch --show-current`
2. Find the base branch (usually `main`): `git merge-base HEAD main`
3. List all changed files: `git diff main --name-only`
4. Find the phase spec directory under `specs/` that matches the current branch/phase.
5. If a phase spec exists, read `plan.md`, `requirements.md`, and `validation.md`.
6. Read `specs/mission.md`, `specs/tech-stack.md`, and `specs/conventions.md`.
7. Find the corresponding epic issue: `gh issue list --label epic --state open --json number,title,body --limit 50` and match by phase name.

## Step 2: Code review

For each changed file, check:

1. **Correctness** — Does the code do what it's supposed to? Are there bugs, edge cases, or error handling gaps?
2. **Go best practices** — Idiomatic Go, proper error handling, no unnecessary complexity.
3. **API design** — Is the public API clean, minimal, and composable? Are types well-named?
4. **Test coverage** — Are the changes adequately tested? Are edge cases covered?
5. **Dependencies** — Are imports minimal and necessary? No operator-controller internals?

## Step 3: Spec consistency

1. If a phase spec exists, verify all requirements in `requirements.md` are addressed.
2. Check that the implementation aligns with `specs/mission.md` design principles.
3. Verify the code follows `specs/tech-stack.md` structure and conventions.
4. Check commit messages follow `specs/conventions.md` format (if any commits exist on the branch).

## Step 4: CLAUDE.md freshness

1. Read `CLAUDE.md` at the project root.
2. Check if any of the following need updating based on changes:
   - Build commands
   - Architecture description
   - Key design decisions
3. If CLAUDE.md needs updates, apply them directly.

## Step 5: Quality gate

1. Run `make check` to verify everything passes.
2. Run `make fmt` and check if it produces any changes.

## Step 6: Report

For issues with multiple valid solutions, use AskUserQuestion to let the user decide. Apply straightforward fixes (formatting, typos, missing error checks) directly.

Summarize:
- Issues found and fixed
- Issues found and flagged for user decision
- Overall assessment (ready to ship or needs more work)

If ready, suggest running `/sdd-ship` to commit and publish.
