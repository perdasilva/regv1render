Brainstorm and create new epics for the project roadmap. Accepts an optional starting idea or theme via $ARGUMENTS.

## Step 1: Gather context

1. Read `specs/mission.md` to understand goals, non-goals, and design principles.
2. Read `specs/tech-stack.md` to understand the project's technical foundation.
3. List existing epic issues (open and closed): `gh issue list --label epic --state all --json number,title,body,state --limit 50`
4. Scan existing spec directories under `specs/` to see what's been planned and implemented: `ls -d specs/*-issue-* 2>/dev/null`
5. Summarize the current state briefly: what exists, what epics are done, what's in progress, what's remaining.

## Step 2: Brainstorm

1. If $ARGUMENTS was provided, use it as a starting point for brainstorming.
2. Consider:
   - Gaps in the current roadmap
   - Unaddressed mission goals
   - Technical debt or quality improvements
   - User-facing features or API improvements
   - Problems or pain points not yet covered
3. Use AskUserQuestion to brainstorm with the user:
   - What problems or gaps are they thinking about?
   - Are there user requests or external drivers?
   - Any technical debt to address?
   - What would make the library more useful?

## Step 3: Propose epics

1. Draft candidate epics, each with:
   - A short name
   - One-line description
   - 3-6 deliverable bullets
   - Why it matters (connection to mission goals)
2. Present the candidates to the user via AskUserQuestion.
3. Iterate: split large epics, merge small ones, reorder by priority, check for dependencies, verify none conflict with non-goals.

## Step 4: Create epic issues

Once the user approves the epics:

1. For each new epic, create a GitHub issue:
   ```
   gh issue create --title "[epic] <name>" --label "epic,ready" --body "<body with deliverables and dependencies>"
   ```
2. Use the same body format as existing epic issues:
   - `## <name>` heading
   - Deliverable bullets
   - `**Dependencies:** #N` section if applicable

## Step 5: Summary

Summarize what was added and suggest running `/sdd-plan-next-phase` to start the next one.
