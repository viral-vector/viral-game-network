# Viral Game Network

Viral Game Network is a game backend written in Go. It manages authentication,
lobbies, and game server lifecycles using Fiber, SurrealDB, Redis, and Kubernetes.
The administration interface uses Pug templates and Stimulus controllers.

## Project layout

- `src/`: backend, API handlers, repositories, scheduled jobs, and Kubernetes integration.
- `assets/`: admin JavaScript, styles, and browser tests.
- `views/` and `public/`: templates and static assets.
- `docker/`: production and development container definitions.
- `tests/`: shared test fixtures.

Operations commands and local tooling documentation live in [bin/README.md](bin/README.md).
That directory is maintained separately and is available only in local checkouts that include it.

## Configuration

Copy [.env.example](.env.example) to `.env` and replace the password and signing-key
placeholders before running the application. The example uses local Docker Compose
service addresses; adjust them when running outside that network. `.env` is ignored
by Git and excluded from Docker build contexts.

Set `VNET_TOKEN_KEY` to a private random value of at least 32 bytes, distinct from
`VNET_KEY`, which is shared as an API key. Use the same signing key across backend
replicas and jobs; never pass it to game clients or game-server pods. Existing
installations need this new variable, and old sessions must sign in again.

Guest login (`POST /auth/guest`) creates a new player identity; display names may
repeat. To renew that identity, send a valid `Viral-Game-Network-Token`, an app key,
and a JSON body to `POST /auth/lobby` before the token expires. That endpoint
refreshes the current player session; a supplied name cannot select another user.
Admin cookies and player tokens have separate audiences and use stable record IDs.
Admin cookies are HttpOnly; login and refresh responses expose expiry metadata
(`expires_at`) instead of the session token.

Private joins accept a JSON body such as `{"code":"your-invitation-code"}` at
`POST /api/lobby/:id/join`. Only the host can retrieve the invitation code from
the lobby API. Repeating an accepted join preserves membership and host ownership.
Lobby patches support explicit `false` and empty values without changing record
IDs, GUIDs, or creation dates.

Provisioned game pods receive `VNET_SERVER_TOKEN`, a token bound to their lobby.
Use it in `Viral-Game-Network-Token` for `GET /api/host/:id/tick`. Renew it with
`POST /api/host/:id/refresh` before expiry, and use the returned `access_token` for
subsequent requests. Other player tokens can heartbeat only the lobby they host.
Server tokens cannot authorize player or admin APIs.

The project is not deployed, so schema changes are folded into the existing
migrations. Use a fresh development database to apply revised initial schemas.

After changing admin source files, run `npm ci --prefix assets` and
`npm --prefix assets run build` before running the Go app directly. The production
Dockerfile builds the admin assets automatically. Kubernetes uses its service
account configuration when running inside a cluster, or an explicit `KUBECONFIG`
when provided; local k3d configuration is the development fallback.
Port allocation requires permission to list nodes and pods across namespaces so
other applications' host-port reservations are respected.

## License

Project code is licensed under the custom [Viral Game Network Use-Only License](LICENSE),
copyright (c) 2026 Viral Vector.

You may use and modify it for your own personal, organizational, or commercial use.
You may not redistribute the original or modified software, source code, binaries,
packages, or container images without prior written permission from Viral Vector,
even if you provide credit. You must retain notices and must not claim the original
work as your own. See LICENSE for the full terms.

This is a source-available license with redistribution restrictions.
Third-party dependencies and files carrying their own license notices remain subject
to those licenses.

## Tests

Run `make test` for Go, admin, and optional local prototype tests. Run `make test-all` to include isolated database integration tests. See [TESTING.md](TESTING.md) for setup, coverage, and CI details. IOC is excluded.
