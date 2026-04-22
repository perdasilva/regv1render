Plan the next phase of work. Find the next epic issue, create a branch, and write a detailed spec.

## Step 1: Pre-flight checks

1. Run `git status`. If there are uncommitted changes, use AskUserQuestion to ask whether to stash them or abort.
2. Check out `main` and pull latest: `git checkout main && git pull origin main`.

## Step 2: Find the next phase

1. List open epic issues that have the `ready` label: `gh issue list --label epic,ready --state open --json number,title,body --limit 50`
2. List all epic issues (open and closed) to check for dependency references: `gh issue list --label epic --state all --json number,title,body,state --limit 50`
3. For each open+ready issue, check if its body contains a "Dependencies:" line referencing other issues. Parse issue numbers from patterns like `Phase N`, `#N`, or issue URLs.
4. An issue is eligible if it has the `ready` label AND all its dependency issues are closed.
5. Among eligible issues, pick the one with the lowest issue number.
6. If no ready issues exist, report the situation and use AskUserQuestion to ask the user what to do.
7. Assign the issue to the current user: `gh issue edit <number> --add-assignee @me`

## Step 3: Create branch

1. Determine a branch name from the issue title. Use format `feat/<short-description>` (e.g., issue "[epic] Project scaffolding" → `feat/project-scaffolding`).
2. Create and check out the branch: `git checkout -b <branch-name>`

## Step 4: Write the phase spec

1. Determine the phase number from the issue number or title.
2. Create a spec directory: `specs/YYYY-MM-DD-phase-N-short-name/` (use today's date).
3. Use AskUserQuestion to ask the user about:
   - What approach to take for this phase
   - Any constraints or decisions to be aware of
   - How to break the work into task groups
   - Validation criteria (how do we know it's done?)
4. Read `specs/mission.md` and `specs/tech-stack.md` for guidance on principles and tech choices.
5. Read the issue body for deliverables.

## Step 5: Write spec files

Create three files in the spec directory:

### plan.md
- Phase title and objective (one paragraph)
- Task groups: ordered groups of related work items
- Each task group has a name, description, and list of specific tasks
- Include estimated complexity (small/medium/large) per task group

### requirements.md
- Functional requirements (what must the code do)
- Non-functional requirements (performance, API design, compatibility)
- Constraints (what we must NOT do or change)
- Dependencies on other code or systems

### validation.md
- Acceptance criteria (specific, testable statements)
- Test scenarios to verify
- Quality gates that must pass (e.g., `make check`)
- How to manually verify the work

## Step 6: Review

After writing all spec files, review them for:
- Internal consistency across the three files
- Consistency with specs/mission.md and specs/tech-stack.md
- Completeness — are all issue deliverables covered?
- Feasibility — are tasks appropriately scoped?

Fix any issues found. Summarize what was created and suggest running `/sdd-implement` to start implementation.
