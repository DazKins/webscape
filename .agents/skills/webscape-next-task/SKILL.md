---
name: webscape-next-task
description: Pick the first Not started Webscape ticket, create a wt worktree, get plan approval, implement it, and open a PR with a Docker preview through Tailscale. Refresh the preview after revisions and merge after feature acceptance. Use for the simple next-task workflow.
---

# Webscape next task

Repository: `/home/dazkins/webscape`

Board: https://app.notion.com/p/3197885724e780d08e9ffc4c43e2cf54

Board view: `view://31978857-24e7-8035-a8e0-000cbba61c00`

1. Fetch the board with the Notion tools, then query its saved view using `notion_query_data_sources` with `mode: "view"` and the view URL above. Pick the first row whose `Status` is `Not started`, preserving returned order. Follow pagination if needed. If there are none, tell the user and stop.
2. Create and open a fresh worktree with the `wt` CLI: from `/home/dazkins/webscape`, run `wt switch --create <unique-task-based-branch-name>`. Derive a concise, meaningful branch name from the selected ticket's title. Resolve the new worktree's path and use it as the working directory for all subsequent repository commands; do not assume directory switching persists between tool calls. If creation fails, resolve that before beginning implementation.
3. Read the ticket, set its `Status` to `In progress`, and tell the user which ticket you picked, including its link.
4. Read the repository instructions and relevant code in the new worktree. Make a short, concrete implementation plan with the checks you will run.
5. Present the plan and ask the user to approve it. Wait for explicit approval before changing implementation files. If the user requests plan changes, revise the plan and ask again.
6. After plan approval, implement that ticket in the worktree and run the relevant checks there.
7. Once the code is ready for review, commit the ticket's changes, push the feature branch to `origin`, and open a GitHub PR targeting the repository's default branch (`master` currently; use `main` if that becomes the default). Include a concise description, validation results, and the Notion ticket link. These actions are authorized by this workflow; do not ask for separate approval. Follow [PR previews](references/pr-previews.md) to wait for the commit's image, run it locally, expose it with Tailscale Serve, and comment on the PR with the verified preview URL and revision. Preview hosting and these PR comments are authorized as part of this workflow. Briefly report what changed and the results, including the PR link, preview link, worktree path, and branch. Make subsequent revisions in the same worktree, run the relevant checks, and commit and push them as additional commits on the same PR. Refresh the preview and its comment for every pushed revision; keep the PR description current. If image publication or hosting fails, report the specific blocker and do not present an older preview as the current revision.
8. When the user approves and accepts the completed feature (for example, "lgtm, feature approved"), squash-merge the PR through GitHub so it adds one commit to the default branch. Feature acceptance authorizes the PR merge; do not ask for separate merge approval. Initial plan approval is not feature acceptance. Preserve unrelated local changes, resolve routine merge conflicts in the feature worktree, and run any checks warranted by conflict resolutions before pushing additional commits to the PR. Respect required checks and branch protections. After a successful merge, set the selected Notion ticket's `Status` to `Done`, stop only this PR's preview using the cleanup instructions in [PR previews](references/pr-previews.md), and mark its preview comment as stopped. If the feature is explicitly abandoned or the PR closes without merging, follow the same ownership-checked preview cleanup without marking the ticket Done. A request for revisions is not abandonment. If completion is blocked, report the specific blocker. Stop after this one ticket.

All implementation work, including investigation, edits, builds, tests, and follow-up revisions, must remain in this same worktree until the feature is accepted by the user and its PR is merged. After the PR merge, use the original checkout as needed to synchronize the default branch while preserving unrelated local changes. Do not remove the feature worktree before the PR merge succeeds.

Keep this workflow simple: use the conversation, PR comment, and labeled Docker containers for continuity, with no state files, locks, checkpoints, delegation, extra approval stages, or production deployment. Commit and push review-ready changes and subsequent revisions automatically; merge the PR only after feature acceptance as described above.
