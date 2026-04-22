Review all changes on the current branch for quality, consistency, and spec compliance.

## Step 1: Gather context

1. Identify the current branch: `git branch --show-current`
2. Find the base branch (usually `main`): `git merge-base HEAD main`
3. List all changed files: `git diff main --name-only`
4. Find the spec directory under `specs/` that matches the current branch.
5. If a spec exists, read `plan.md`, `requirements.md`, and `validation.md`.
6. Read `specs/mission.md`, `specs/tech-stack.md`, and `specs/conventions.md`.
7. Find the corresponding epic issue: `gh issue list --label epic --state open --json number,title,body --limit 50` and match by branch name.

## Step 2: Code review

For each changed file, check:

1. **Correctness** — Does the code do what it's supposed to? Are there bugs, edge cases, or error handling gaps?
2. **Go best practices** — Idiomatic Go, proper error handling, no unnecessary complexity.
3. **API design** — Is the public API clean, minimal, and composable? Are types well-named?
4. **Test coverage** — Are the changes adequately tested? Are edge cases covered?
5. **Dependencies** — Are imports minimal and necessary? No operator-controller internals?

## Step 3: Spec consistency

1. If a spec exists, verify all requirements in `requirements.md` are addressed.
2. Check that the implementation aligns with `specs/mission.md` design principles.
3. Verify the code follows `specs/tech-stack.md` structure and conventions.
4. Check commit messages follow `specs/conventions.md` format (if any commits exist on the branch).

## Step 4: Scope drift

Compare what was actually implemented against the spec and the epic issue to detect scope changes — work that was added, removed, or changed during implementation.

1. Read the epic issue body (`gh issue view <number> --json body`) and compare its deliverables against the actual changes on the branch.
2. Read `plan.md` task groups and check whether any tasks were added, skipped, or significantly changed during implementation.
3. Read `requirements.md` and check whether any requirements were added or dropped.
4. For each scope change found, classify it:
   - **Added scope** — work done that wasn't in the original plan (e.g., removed a feature, consolidated packages, added a convenience API)
   - **Dropped scope** — planned work that wasn't done
   - **Changed approach** — work done differently than planned
5. If scope changes are found:
   - Update `plan.md` to reflect what was actually delivered, not what was originally planned
   - Update `requirements.md` and `validation.md` if requirements or acceptance criteria changed
   - Update the epic issue body via `gh issue edit <number> --body` to reflect the actual deliverables
   - Use AskUserQuestion if a scope change is ambiguous or if you're unsure whether a change was intentional

## Step 5: CLAUDE.md freshness

1. Read `CLAUDE.md` at the project root.
2. Check if any of the following need updating based on changes:
   - Build commands
   - Architecture description
   - Key design decisions
3. If CLAUDE.md needs updates, apply them directly.

## Step 6: Quality gate

1. Run `make verify` to verify everything passes.
2. Run `make fmt` and check if it produces any changes.

## Step 7: Report

For issues with multiple valid solutions, use AskUserQuestion to let the user decide. Apply straightforward fixes (formatting, typos, missing error checks) directly.

Summarize:
- Issues found and fixed
- Scope changes detected and updated
- Issues found and flagged for user decision
- Overall assessment (ready to ship or needs more work)

If ready, suggest running `/sdd-ship` to commit and publish.
