# Testing

Use Go 1.24 or newer, Node.js 22, GNU Make, and a C compiler for Go race detection. Docker with Compose is required for the managed integration database. Install admin dependencies with `npm ci --prefix assets`. If the independent local `bin` repository is present, also run `npm ci --prefix bin/proto`.

`make test` runs Go unit tests with race detection, admin DOM tests, and the local prototype tests when available. Unit tests use isolated in-memory Redis instances and fake Kubernetes clients; they never create a cluster. No IOC code is tested.

`make test-all` also runs the integration suite. Docker Compose starts a disposable, memory-only SurrealDB 2.2.2 instance on localhost port 18000, waits for readiness, and removes it afterward. This version matches the existing SDK and migration syntax. Each integration test uses its own randomly named namespace and the fixed test/test credentials. Tests do not read production database credentials.

To use an already running **dedicated test** instance instead of Docker:

```sh
VGN_TEST_STORE_ENDPOINT=ws://127.0.0.1:18000/rpc make test-integration
```

The integration build tag requires that explicit endpoint; a missing endpoint fails rather than silently skipping tests. Test namespaces remain on an externally managed instance until that instance is reset. Never point this endpoint at a production database.

`make coverage` produces Go coverage in `coverage/backend.out`, admin HTML and LCOV reports in `assets/coverage`, and prototype coverage in the terminal. View backend coverage with `go tool cover -html=coverage/backend.out`.

## Coverage areas

- Backend: JWT and password authentication, Redis cache and pub/sub, lobby notifications, model serialization, form generation, merge semantics, migrations and repositories, HTTP authentication boundaries, bootstrap and admin views, and server provisioning, status, cleanup and lobby expiry.
- Kubernetes: pod configuration, labels, placement, port selection and exhaustion, namespace creation, command failures, and missing cluster configuration. Clients and k3d commands are faked; real cluster deployment is outside this suite.
- Admin: actual Stimulus DOM bindings for forms, authentication refresh, idle timers, navigation, search, pagination, dialogs, progress, notifications, API keys and chart initialization.
- Local prototype: HTTP status response, guest authentication and heartbeat contract, token renewal, network failures, overlapping requests, and timer shutdown. These tests live in `bin/proto/test` and are tracked by the independent `bin` repository.

GitHub Actions runs backend integration tests with race detection and admin tests plus the production asset build. The ignored local `bin` repository and IOC are absent from the parent checkout and its CI. This is a starting regression suite, not a claim of complete line or branch coverage.
