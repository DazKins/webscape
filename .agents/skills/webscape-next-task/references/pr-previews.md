# PR previews

Use this after opening a PR and after every subsequent push. Run commands from the
ticket's existing worktree. GitHub Actions builds the image; the local agent hosts
it until the feature is accepted and the PR merges. There is no hosting daemon or
production deployment step.

## Identify the image

Read the PR number, head SHA, and repository with `gh pr view` and `gh repo view`.
Require a clean worktree whose `git rev-parse HEAD` equals the PR head SHA. Wait for
the `Build PR preview image` workflow run associated with that PR and revision to
succeed, inspecting its event and PR association as well as its SHA. Do not select
an unrelated run or an older successful build. Bound the wait to the workflow's
30-minute timeout, then report a missing, failed, or timed-out run.

Pull `ghcr.io/<lowercase-owner>/<lowercase-repo>:pr-<number>-sha-<full-head-sha>`.
Use `docker`, or `sudo -n docker` if that is how this host grants Docker access.
Inspect the pulled image's `org.opencontainers.image.revision` and
`io.webscape.prerelease` labels; they must equal the PR head SHA and `true`.
Run the pulled image by its immutable image ID or digest.

These are testing-only prerelease images. GHCR has no GitHub Release-style
prerelease switch: separation comes from the `pr-...-sha-...` tag and metadata.
Never publish or run a preview under `latest`, a production `sha-...` tag, or a
release version tag. Do not create a GitHub Release for a preview. The production
workflow continues to own its existing tags.

Fork and Dependabot PRs build without publishing because their workflow tokens
cannot write packages. Report that limitation if such a PR needs a preview;
do not switch to `pull_request_target`, inject credentials, or substitute a
production image. Likewise, report unavailable Docker, Tailscale, or GHCR access.

## Start an isolated preview

1. Inspect existing Docker containers and `tailscale serve status --json`. Use a
   container name containing the repository, PR number, and revision. Label it
   with `io.webscape.preview.repository=<owner/repo>`,
   `io.webscape.preview.pr=<number>`,
   `io.webscape.preview.revision=<full-head-sha>`, and
   `io.webscape.preview.worktree=<absolute-worktree-path>`. Reuse a healthy preview
   only if all labels and its image ID match. Never take over a conflicting name.
2. Start the image detached with `--publish 127.0.0.1::8080`. Let Docker allocate
   the port, then read the actual mapping with `docker port <container> 8080/tcp`.
   Do not mount the worktree, credentials, or Docker socket into the container.
   Check the container stays running and HTTP `/` returns the game client before
   exposing it. Allow up to 60 seconds for startup; inspect logs on failure.
3. Use that allocated port as the preview's HTTPS port too. Each running container
   reserves its own loopback port, so concurrent agents following this procedure
   get distinct ports without a shared lock or state file. Immediately re-read
   `tailscale serve status --json` and check host listeners. If that port already
   has a Tailscale handler, another listener outside this container's mapping, or
   a Funnel configuration, remove only the new container and retry allocation
   (at most three attempts). Never overwrite an existing endpoint.
4. Run `tailscale serve --bg --https=<allocated-port> http://127.0.0.1:<allocated-port>`
   (with `sudo -n` if needed). Use Serve, not Funnel. Read the node's DNS name from
   `tailscale status --json` and verify that the new handler maps only this HTTPS
   port's `/` to this container's loopback address. Other handlers must remain
   unchanged. The preview URL is `https://<node-dns-name>:<allocated-port>/` and
   requires access to the tailnet.
5. Check HTTP and a WebSocket connection to `/ws` through that HTTPS URL, using
   normal certificate verification, and confirm the client build revision. The
   dedicated port lets root-relative assets and `/ws` work without path rewriting.
   Re-read the PR head before reporting success; if it changed, refresh to the
   new revision. On failure, clean up only this attempted preview and report the
   problem. Keep any previous working preview until the replacement passes.

## Publish and refresh the link

Post a PR comment containing the marker `<!-- webscape-pr-preview -->`, preview
URL, full source SHA, image tag and digest/ID, container name, worktree path, and
the note "Testing-only prerelease preview; tailnet access required." Posting this
comment is part of the authorized task workflow. Prefer editing this same comment
on subsequent revisions. Identify it by both the marker and the authenticated
author; do not use `--edit-last` on an arbitrary comment. Use a structured API body
or a temporary Markdown file with `gh pr comment --body-file` to preserve newlines.

After each push, mark the comment as updating and identify the old revision until
the new image passes the checks. Start the replacement alongside the old preview,
update the comment to the new URL/revision, then clean up the old preview. The URL
may change between revisions. If the update fails, say so on the same comment,
clearly identifying any still-running older revision. Keep the active preview
running while awaiting feature acceptance; do not merge based on plan approval.

## Cleanup

For each old, failed, merged, or explicitly abandoned preview, inspect its container labels and port
mapping again. Require the repository, PR, and worktree to match this task. Before
disabling Serve, verify that the endpoint contains only the expected `/` proxy to
that exact loopback port and has no unrelated TCP handler or Funnel configuration.
Then run `tailscale serve --https=<port> off`, followed by
`docker rm -f <exact-owned-container>`. If ownership or routing has changed, leave
it intact and report the conflict. If Serve was never enabled and the port is
unused by Tailscale, remove just the owned container.

Never use `tailscale serve reset`, bulk container deletion, or Docker prune. Do not
delete images or GHCR tags as part of cleanup. On successful PR merge, mark the
preview comment as stopped and retain the reviewed revision and image reference.
If the merge fails, keep the active preview available.

When the user explicitly cancels/abandons the feature, or when resuming work reveals
that the PR is closed without a merge, clean up its owned preview the same way and
mark the comment as stopped because the feature was abandoned or closed unmerged.
Do not mark the Notion ticket Done in that case. Requests for fixes or more review
keep the preview available; they do not count as abandonment. Cleanup runs when
the agent handles the cancellation or observes the closed PR, not automatically
while no agent is active.
