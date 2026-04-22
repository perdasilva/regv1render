Verify, commit, and publish the current epic's work.

## Step 1: Verify

1. Run the full quality gate: `make verify`
2. Find the spec directory under `specs/` that matches the current branch.
3. If a spec exists, walk through `validation.md` acceptance criteria and verify each is met.
4. Check CLAUDE.md freshness — verify build commands, architecture summary, directory structure, and conventions all reflect the current state.
5. Run `make fmt` and check for uncommitted formatting changes.
6. If any verification fails, fix it before proceeding.

## Step 2: Commit

1. Read `specs/conventions.md` for commit message and PR format.
2. Run `git status` to see all changes.
3. Run `git diff --stat` to summarize the scope of changes.
4. Draft a commit message following conventional commits format.
5. Use AskUserQuestion to show the user the proposed commit message and ask for confirmation before committing.
6. If there are multiple logical changes, ask whether to split into separate commits or keep as one.
7. Stage and commit the changes.
8. If there were previous review-fix commits on this branch, ask whether to squash them into a clean commit history.

## Step 3: Publish

1. Use AskUserQuestion to confirm before pushing.
2. Push the branch: `git push -u origin <branch-name>`
3. Find the corresponding epic issue number. Search open epic issues: `gh issue list --label epic --state open --json number,title --limit 50` and match by branch name.
4. Create a PR using `gh pr create`:
   - Title: matches the primary commit subject (conventional commit format)
   - Body format:

```markdown
## Summary
- <bullet points describing what changed and why>

## Test Plan
- [ ] <verification steps>

Closes #<epic-issue-number>
```

5. Report the PR URL to the user.
