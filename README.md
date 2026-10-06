# Viral Game Network

Viral Game Network is a game backend for managing players, multiplayer lobbies,
and game-server pods. It includes a web administration panel and uses Kubernetes
to provision game servers, SurrealDB for storage, and Redis for messaging.
See [LICENSE](LICENSE) for permitted use.

## Features

- Guest player sessions and a separate administrator login.
- Public and private lobbies with invitation codes and player limits.
- Game-server provisioning, heartbeat readiness, and lifecycle cleanup.
- Lobby chat and live status notifications.
- Web administration for applications, players, lobbies, servers, and settings.

## Technologies and why they fit

| Technology | Role and rationale |
| --- | --- |
| Go and Fiber | Power the HTTP API, administration panel, and background workers. Go's concurrency model fits handling player connections alongside game-server lifecycle work. |
| Kubernetes and [K3s](https://docs.k3s.io/) | Kubernetes runs game servers as pods and provides scheduling and lifecycle management. K3s is a lightweight Kubernetes distribution that makes smaller deployments practical; the backend uses the standard Kubernetes API. |
| [SurrealDB](https://surrealdb.com/docs/learn/data-models/graph/overview) | Stores players, games, lobbies, servers, and configuration. Its graph relationships fit lobby membership and the links between a lobby, its game, and its server. |
| Redis | Provides caching, chat history, publish/subscribe messaging, rate limits, and shared locks. These support live updates and coordination between backend workers. |
| Pug, Stimulus, and Bulma | Pug renders the administration pages, Stimulus adds browser interactions, and Bulma supplies the layout and styling. Together they support a responsive interface built around server-rendered HTML. |
| Docker | Packages the backend and web assets into one image, giving the web application and scheduled workers a consistent deployment environment. |

## Boot and deploy

You need Docker, reachable Redis and SurrealDB services, and a Kubernetes cluster
for running game servers. SurrealDB 2.2.2 is the version verified with this app.

1. Clone the repository and create your configuration:

   ```sh
   git clone https://github.com/viral-vector/viral-game-network.git
   cd viral-game-network
   cp .env.example .env
   ```

2. Edit `.env`: replace the admin and database passwords, set a random API key,
   and set a separate private `VNET_TOKEN_KEY` containing at least 32 random bytes.
   Configure the Redis and SurrealDB endpoints for your deployment. Set
   `VNET_HOST` to an address your game pods can reach, and leave `VNET_JOB_NAME`
   empty to start the web application. Keep `.env` private.

3. Build the application image:

   ```sh
   docker build -f docker/Dockerfile -t viral-game-network .
   ```

4. Run it with your configured services:

   ```sh
   docker run --rm --name viral-game-network --env-file .env \
     -p 3000:3000 viral-game-network
   ```

   The service addresses in `.env` must be reachable from the container. The
   example uses `cache` and `store` as hostnames; replace them or attach the
   container to the Docker network containing those services.

5. Open `http://localhost:3000/auth/admin` and sign in with the admin credentials
   configured in `.env`. Register your game applications through the admin panel.

Schedule recurring workers using the same image and service configuration, with
`VNET_JOB_NAME` set to `monitoring`, `lobby_server_provisioner`,
`lobby_server_stewardship`, or `lobby_stewardship`. Each worker invocation runs
once and exits; the web container does not run them automatically.

For game-server provisioning, create the `vgn-app` namespace and give the backend
Kubernetes credentials that can manage game pods and read nodes and pods across
namespaces. Run the image in Kubernetes with a service account, or mount a readable
kubeconfig into the container and set `KUBECONFIG` to its path. Expose public
installations through an HTTPS ingress or reverse proxy, and persist your database
and Redis data in the services hosting them.
