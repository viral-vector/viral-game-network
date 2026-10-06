# Viral Game Network

Viral Game Network is a game backend written in Go. It manages authentication,
lobbies, and game server lifecycles using Fiber, SurrealDB, Redis, and Kubernetes.
The administration interface uses Pug templates and Stimulus controllers.

## Project layout

- `src/`: backend, API handlers, repositories, scheduled jobs, and Kubernetes integration.
- `assets/`: admin JavaScript and styles.
- `views/` and `public/`: templates and static assets.
- `docker/`: production and development container definitions.

Operations commands and local tooling documentation live in [bin/README.md](bin/README.md).
That directory is maintained separately and is available only in local checkouts that include it.

## Configuration

Copy [.env.example](.env.example) to `.env` and replace the password and signing-key
placeholders before running the application. The example uses local Docker Compose
service addresses; adjust them when running outside that network. `.env` is ignored
by Git and excluded from Docker build contexts.

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
