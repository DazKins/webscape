# Authentication

Webscape uses a configurable OpenID Connect (OIDC) provider. The provider runs the
login page and owns passwords, signup and account recovery. Webscape is a
confidential web client in the default OIDC mode; it does not implement those account-management features.

## Configure a provider

Register a web application with your provider. It must support OIDC discovery,
authorization code flow, PKCE S256, signed ID tokens through a JWKS endpoint, and
`client_secret_basic` or `client_secret_post` token endpoint authentication.
Webscape requests only `openid`; email, names and vendor-specific claims are not
required. OAuth-only providers need an OIDC broker or a separate adapter.

Set these fields in the deployment's root `config.json`:

```json
"auth": {
  "issuer": "https://identity.example.com",
  "clientId": "webscape",
  "clientSecretEnv": "WEBSCAPE_OIDC_CLIENT_SECRET",
  "publicUrl": "https://game.example.com",
  "sessionLifetimeSeconds": 28800
}
```

Replace the example URLs/client ID. Set the named environment variable to the
provider-issued client secret through your deployment's secret settings. Never
commit the secret or include it in a URL. Missing auth settings or credentials,
failed discovery and unsupported configuration prevent server startup. The
checked-in example deliberately requires configuration; there is no automatic guest fallback.

Allow exactly `https://game.example.com/auth/callback` as a callback at the
provider. `publicUrl` is the browser-facing origin, without a path, query or
fragment. It is independent of `server.address`, which can remain `:8080` behind
an HTTPS reverse proxy. The server trusts this configured origin rather than
forwarded headers. The proxy must support WebSocket upgrades and forward cookies
and the browser's Origin header. Do not expose an additional unencrypted public
entry point to the application.

Use the same client assets in every environment; no auth credentials or issuer
are embedded by Vite. For Docker, mount your runtime config read-only at
`/app/config.json` and inject the named secret environment variable through your
container/deployment secret settings. The existing production deployment workflow
does not create a provider or configure secrets for you.

## Local development and previews

For local HTTP development, set `server.devMode` to `true` and use loopback HTTP
URLs (for example `http://127.0.0.1:8080` and
`http://127.0.0.1:5556/dex`). Configure the provider's exact loopback callback URL.
Use a compatible local provider such as Dex, or your hosted provider's development
application. HTTP is rejected for non-loopback auth URLs even in development;
OIDC mode always verifies identity. Use the Go server to serve
the built client, so auth routes, cookies and `/ws` share one origin.

For a tailnet preview, set `publicUrl` to that preview's HTTPS origin (including
its allocated port), register its `/auth/callback`, and supply a test provider
secret to the runtime securely. A preview without a configured provider cannot
start. Do not treat a passing unauthenticated handshake as success: `/ws` should
return 401 with a valid Origin and no session, then upgrade only after login.

## No-auth testing mode

To test without any provider, explicitly set `server.devMode` to `true` and use:

```json
"auth": {
  "mode": "none",
  "publicUrl": "http://127.0.0.1:8080",
  "sessionLifetimeSeconds": 28800
}
```

No issuer, client ID or secret is needed in this mode. Use the actual HTTPS origin
for a hosted testing preview. The client shows **Play as guest**, followed by name
entry. Anyone who can access this server can play; this mode is for testing.

The server assigns a random guest character ID and keeps it in a cookie-backed
session. Reload/reconnect keeps that ID while the session remains valid. Ending
the session, expiry, clearing cookies or restarting the server loses access to
that guest character; a new session starts with a new ID. Guest IDs cannot claim
OIDC characters or old anonymous saves. With persistence enabled, guest snapshots
may remain stored but are not recoverable after the session ends; prefer a
separate test world with `persistence.driver: "none"`.

Origin checks, CSRF-protected session termination, socket expiry and server-side
command validation still apply. Guest cookies are separate from OIDC cookies.
Omitting `auth.mode` means `oidc`; unknown modes, missing OIDC credentials and
provider discovery errors fail startup rather than enabling guest play.

## Sessions, logout and saved characters

The server verifies token signature, issuer, audience, expiry and nonce, and
consumes each browser-bound authorization transaction once. The five-minute
login transaction uses PKCE. Provider errors redirect to a generic retry screen
without logging codes, tokens or secrets.

SCS keeps sessions in memory. Browsers receive an HttpOnly, SameSite=Lax cookie,
with Secure and the `__Host-` prefix for HTTPS; loopback HTTP uses a separate local
cookie name. Session lifetime is 60–86400 seconds and is capped by ID-token expiry.
No refresh tokens are requested. Sessions end on restart. Reauthentication can be
silent at a provider that retains an SSO session.

All game sockets require a session and the exact configured Origin. Logout is a
same-origin POST with a session CSRF token. It invalidates that browser session
and closes all its sockets; other browsers have independent sessions. Session
expiry also closes idle sockets. Commands and outgoing updates check session
validity. Sign out ends the Webscape session, not the provider's global SSO
session; provider-wide logout is deliberately not a vendor-specific dependency.

Characters are keyed by a fixed SHA-256-derived UUID from the verified issuer and
subject. Neither a display name nor browser local storage proves ownership.
Registration sends only `{ "name": "Adventurer" }`; supplying `id` is rejected.
The browser may remember the display name separately for each account.

With PostgreSQL persistence enabled, the same issuer/subject recovers character
progress across browsers and server restarts. `persistence.driver: "none"` still
means ephemeral game progress. Old anonymous saves are preserved but cannot be
claimed by submitting their UUID or name. New authenticated accounts start fresh;
linking anonymous saves needs a separately verified administrative migration.
Changing provider/issuer or a provider's subject configuration also needs an
explicit migration. The existing one-active-connection-per-character rule still
applies. Account deletion, provider-wide revocation notifications and automatic
identity linking are not implemented; local sessions remain bounded by expiry.

## Validation

`go test ./...` exercises mock-provider code exchange, token validation,
transaction replay, cookie and Origin checks, account isolation, socket logout,
and malformed commands. With `WEBSCAPE_TEST_DATABASE_URL` pointing to an isolated
PostgreSQL database, it also exercises authenticated progress across restart and
reconnect. Run `go test -race ./server ./server/auth ./server/config`,
`go build ./...`, and `cd client && pnpm run build` for the complete auth checks.
The mock provider in `server/internal/oidctest` is test support, not a runtime login
option. Browser validation should also use an actual configured OIDC provider.

References: [OIDC Core](https://openid.net/specs/openid-connect-core-1_0.html),
[OIDC Discovery](https://openid.net/specs/openid-connect-discovery-1_0.html),
[go-oidc](https://github.com/coreos/go-oidc),
[SCS](https://github.com/alexedwards/scs).
