# Deployment

## Published Docker image
The `Publish Docker image` GitHub Actions workflow builds the root Dockerfile on
pushes to `master`, including merged pull requests, and publishes a Linux AMD64
image to GitHub Container Registry:

- `ghcr.io/dazkins/webscape:latest` tracks the latest successful publication.
- `ghcr.io/dazkins/webscape:sha-<full-commit-sha>` identifies a specific revision.

The workflow passes the source SHA into the client build identifier and image
metadata. It authenticates with the automatic `GITHUB_TOKEN` using
`contents: read` and `packages: write`; no additional registry secret is needed.
After successful publication, the workflow requests a Coolify deployment of the
new `latest` image. A failed build or publication never triggers deployment.

First configure `config.json` and export `WEBSCAPE_OIDC_CLIENT_SECRET` as described
in [Authentication](authentication.md). Then pull and run a published image
behind your HTTPS reverse proxy:

```sh
docker pull ghcr.io/dazkins/webscape:latest
docker run --rm -p 127.0.0.1:8080:8080 \
  --mount type=bind,src="$(pwd)/config.json",dst=/app/config.json,readonly \
  --env WEBSCAPE_OIDC_CLIENT_SECRET \
  ghcr.io/dazkins/webscape:latest
```

If `auth.clientSecretEnv` uses a different variable name, export and pass that name
instead. The mounted config must be readable by the container user.

For a specific revision, replace `latest` with its `sha-<full-commit-sha>` tag.
Package visibility is managed in GitHub Packages settings; make the package public
if anonymous pulls are required, or authenticate to GHCR before pulling a private
package.

## Coolify auto-deploy
Configure the Coolify application to deploy the Docker image
`ghcr.io/dazkins/webscape:latest`. In Coolify, create an API token under
**Keys & Tokens → API Tokens** with the **Deploy** permission, and copy the
application's **Webhook → Deploy webhook** URL. See the
[Coolify GitHub Actions guide](https://coolify.io/docs/applications/ci-cd/github/actions/).

Add these repository secrets under GitHub **Settings → Secrets and variables →
Actions**:

- `COOLIFY_WEBHOOK`: the application's HTTPS deploy webhook URL, reachable from
  GitHub-hosted runners.
- `COOLIFY_TOKEN`: the API token with Deploy permission.

The workflow sends an authenticated HTTPS `POST` to `/api/v1/deploy`, preserving
the application UUID and other query parameters from the copied webhook URL.

Only the `master` publication workflow calls this webhook, after the image push
succeeds. PR preview builds publish their separate prerelease images and never
request a Coolify deployment. Publication and the webhook request share the same
workflow concurrency group.

Missing secrets, network errors, and non-2xx responses fail the workflow's
deployment step even though the image has already been published. The request
has a 60-second timeout and is not automatically retried: a timeout can occur
after Coolify has queued a deployment. Check Coolify before retrying. A successful
request means Coolify accepted it; monitor the rollout in Coolify to confirm the
game is running the expected revision. Keep tokens in GitHub secrets, never in
source control.
