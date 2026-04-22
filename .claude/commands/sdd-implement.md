Implement the current epic based on its spec. Follow the plan, validate as you go.

## Step 1: Load context

1. Identify the current branch: `git branch --show-current`
2. Find the spec directory under `specs/` that matches the current branch. Look for the most recent `specs/*-issue-*` directory.
3. Read all three spec files:
   - `plan.md` — task groups and order of work
   - `requirements.md` — what must be built and constraints
   - `validation.md` — acceptance criteria and quality gates
4. Read `specs/mission.md` and `specs/tech-stack.md` for project-level guidance.

## Step 2: Implement task groups

For each task group in `plan.md`, in order:

1. Announce which task group you're starting.
2. Implement each task in the group.
3. After completing the group, run `make check` (or the available subset if early epics haven't set up all targets yet).
4. Fix any issues before moving to the next group.
5. If a decision isn't covered by the spec, use AskUserQuestion to ask the user.

## Step 3: Validate

After all task groups are complete:

1. Run the full quality gate: `make check`
2. Walk through each acceptance criterion in `validation.md` and verify it's met.
3. Run any manual verification steps described in `validation.md`.
4. If any criterion is not met, fix the issue and re-validate.

## Step 4: Summary

Report what was implemented, what tests pass, and any decisions made during implementation. Suggest running `/sdd-review` to review the changes before shipping.
