# Load Balancer

A TCP load balancer with round-robin, least-connections, and random strategies. Supports runtime configuration via a management client.

## Build

```bash
go build -o balancer ./cmd/balancer
go build -o client ./cmd/client
```

## Usage

A config file is required:

```bash
./balancer --config example.config.yaml
```

### Balancer flags

All flags are optional and override config file values:

| Flag | Description |
|---|---|
| `--config PATH` | Path to config file (required) |
| `--listener-port PORT` | Port for incoming connections |
| `--client-port PORT` | Port for the management API |
| `--strategy NAME` | `roundrobin`, `leastconnections`, or `random` |
| `--server-conn-timeout SEC` | Backend connection timeout |
| `--timeout SEC` | Idle connection timeout |
| `--max-connections N` | Max concurrent connections |
| `--backends ADDR PORT MAX [...]` | Backend list (repeatable triples) |

### Config file

```yaml
ListenerPort: 8080
ClientPort: 9080
LoadBalancingStrategy: RoundRobin
ServerConnTimeoutSec: 30
IdleTimeoutSec: 120
MaxConnections: 1000
Backends:
  - Address: "127.0.0.1"
    Port: 8081
    MaxConnections: 300
  - Address: "127.0.0.1"
    Port: 8082
    MaxConnections: 300
```

#### Note that every config field has to be specified either through flag or config file.

## Management client

The client connects to the management API to control the balancer at runtime.

```bash
./client --port 9080 <command>
```

### Commands

| Command | Description |
|---|---|
| `status` | Show balancer status |
| `strategy <name>` | Change strategy (`roundrobin`, `leastconnections`, `random`) |
| `backend ls` | List backends |
| `backend conns` | Show connections per backend |
| `backend add <addr> <port> <max>` | Add a backend |
| `backend remove <addr> <port>` | Remove a backend |
| `timeout <seconds>` | Set idle timeout |
| `max-connections <count>` | Set max connections |
| `stop <graceful\|immediate>` | Stop the balancer |
| `start` | Start the balancer |
| `shutdown [timeout_sec]` | Shutdown the application (default: 5s) |

## Testing with example servers

Start dummy HTTP backends:

```bash
python3 example/servers.py 2 8081
```

This starts 2 servers on ports starting from specified 8081 (8081 and 8082). Then run the balancer:

```bash
./balancer --config example.config.yaml
```

Send requests through the load balancer:

```bash
curl http://localhost:8080
```

Each request will be forwarded to a different backend depending on the strategy.
Remember that most of the browsers make more than one TCP session while getting even
simple html from GET request.

## Technical breakdown

The balancer operates at the TCP level — it accepts client connections, picks a backend using the configured strategy, and performs bidirectional `io.Copy` between the client and backend sockets. Each connection runs in its own goroutine.

### Shutdown modes

- **Graceful** (`SIGTERM` or client `stop graceful`) — stops accepting new connections, waits for all active connections to finish.
- **Immediate** (`stop immediate`) — stops accepting and cancels all active connections via context. Can also be triggered during a graceful shutdown to force-close remaining connections.
- **Shutdown** (`shutdown`) — immediately stops the balancer and exits the process.

### Strategies

All strategies implement the `Strategy` interface (`PickBackend` + `OnRelease`). They share a `BackendConnections` tracker so connection counts persist across strategy swaps at runtime.

- **RoundRobin** — cycles through backends sequentially, skipping those at max connections (convenient for testing balancing due to cycles).
- **LeastConnections** — picks the backend with the fewest active connections.
- **Random** — picks a random backend, falling through to the next if at capacity.

### Project structure

```
cmd/balancer/       — balancer entry point
cmd/client/         — management CLI client
internal/balancer/  — core balancing logic, connection handling, idle timeout
internal/strategy/  — strategy interface and implementations
internal/ui/        — HTTP management API
internal/config/    — YAML config loading
internal/flags/     — CLI flag parsing
internal/common/    — shared types (Backend)
example/            — dummy HTTP servers for testing
```