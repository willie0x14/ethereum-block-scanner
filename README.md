# Ethereum Block Scanner

Minimal Ethereum Block Listener written in Go.

This project is a learning-focused implementation that combines:

* Block Listener (ETH RPC / WS)
* RESTful API (net/http or Gin)
* Repository abstraction (Postgres planned)

---

## Project Structure

```
cmd/main.go

internal/
├── node/              # Ethereum client wrapper
├── listener/          # Block subscription / polling logic
├── service/           # Business logic layer
├── repository/        # Repository interface + memory impl (v1)
├── api/               # REST handlers
├── model/             # Event models
└── config/            # Configuration loader
```

---

## v1 Features

* Connect to Ethereum node
* Listen for new blocks
* Process block events (basic version)
* In-memory storage (temporary)
* REST API endpoints:

  * `GET /health`
  * `GET /status`
  * `GET /events?limit=20`

---

## Quick Start

### Prerequisites

### Start the stack
```bash
go mod tidy
go run ./cmd
```

Server default:

```
http://localhost:8080
```

---

## Example API

### Health Check

```
GET /health
```

Response:

```json
{
  "status": "ok"
}
```

---

### Status

```
GET /status
```

Response:

```json
{
  "listener_running": true,
  "last_processed_block": 12345678
}
```

---

## TODO (Future Work)

### Database

*

### Reliability

*

### Performance

*

### Observability

*

### Testing

*

---

## Learning Goals

This project is intentionally designed to practice:

* Clean architecture in Go
* Interface-driven design
* context propagation
* REST API design
* Dependency injection
* Testing with mocks

---

## License

Personal learning project.
