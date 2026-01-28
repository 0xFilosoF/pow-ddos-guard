# DDoS Guard with Proof of Work

A lightweight TCP guard that enforces **Proof-of-Work (PoW)** on every incoming connection before allowing any meaningful interaction.  
The goal is to make large-scale connection floods economically expensive while keeping the server-side logic simple, fast, and stateless.

## Overview

This project demonstrates a **connection-level anti-DDoS mechanism** based on Proof-of-Work.

Each client must solve a cryptographic challenge **per connection**.  
Only after a valid solution is provided, the server starts accepting RPC messages from that client.

### Crypto

**Why Hashcash-style Proof of Work**

This project uses a **Hashcash-style PoW** (leading-zero-bits) rather than memory-hard algorithms like Argon2 or scrypt.

**TL;DR**
> Hashcash is simple, fast to verify, predictable, and well-suited for connection-level protection.

<details>
<summary>Why Hashcash-style</summary>

#### Hashcash (leading-zero-bits)

https://github.com/0xFilosoF/pow-ddos-guard/blob/8e5e663a9463582b12806885fa1ceaa1f94839dc/internal/shared/pow/hashcash.go#L26-L35

A client must find a nonce such that:

Properties:
- Solving cost grows exponentially with difficulty
- Verification is constant-time and extremely cheap
- No server-side state is required
- Easy to adjust difficulty dynamically

This makes Hashcash ideal for:
- TCP / RPC guards
- High-throughput services
- Scenarios where verification must stay cheap under attack

#### Why not Argon2 / scrypt

Memory-hard PoW (Argon2, scrypt, etc.) was intentionally **not** chosen here:

- Requires large memory allocations per client
- Hard to tune safely for heterogeneous clients
- Risk of memory exhaustion under attack
- Slower verification
- Poor fit for low-latency connection handshakes

Such algorithms are better suited for:
- Anti-scraping
- Login protection
- High-value endpoints

But for a **connection gate**, they introduce unnecessary complexity and operational risk.

#### Other challenge/puzzle types

The design is compatible with other PoW variants:
- Sequential hash chains
- Time-based PoW
- Hybrid PoW + rate limiting

However, Hashcash offers the best trade-off between simplicity and effectiveness for this use case.

</details>

### Architecture

The server operates as a **stateless gate**:

1. Client connects via TCP
2. Server sends a PoW challenge
3. Client solves the puzzle and responds
4. Server verifies the solution
5. Only then application-level RPC is allowed

<details>
<summary>TCP PoW Guard Flow</summary>

#### Stateless TCP PoW Guard

Important properties:
- No per-client storage
- No challenge persistence
- No cleanup goroutines
- No timers for state eviction

Each connection is fully independent.  
If the logic changes, every new connection must solve a new challenge.

This avoids common pitfalls such as:
- Storage leaks
- Timer storms
- Goroutine buildup
- State desynchronization under load

#### Why JSON-RPC 2.0 (Notify-based)

All communication is built on **JSON-RPC 2.0**, primarily using **Notify** messages.

Reasons for choosing JSON-RPC:
- Clear message framing over TCP
- Strong separation of protocol vs transport
- Easy extensibility
- Human-readable (useful for debugging)

Why **Notify channels** specifically:
- No request/response state to keep
- No request IDs to track
- No server-side queues
- Naturally fits a fire-and-forget handshake

If the client does not solve the PoW within the allowed time window,  
the connection is closed automatically.

The effective timeout is:

This guarantees:
- No hanging connections
- No leaked resources
- Predictable connection lifetime

</details>

## Development

### Common Commands

- `make dev-server` — run server in dev mode
- `make dev-client` — run client in dev mode
- `make test` — run tests

#### Build
- `make build` — build debug binaries (→ `bin/debug/`)
- `make build-release` — build production binaries for Linux (→ `bin/release/`)

#### Linting
- `make lint` — check code style
- `make lint-fix` — auto-fix code style issues
- `make lint-fast` — quick partial lint check

#### Utilities
- `make clean` — remove binaries and git hooks
- `make install-deps` — install linter and pre-commit hooks

### Deploy
```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### TLS Configuration

The TCP server uses TLS for secure connections. Ensure certificates are properly configured before running.


