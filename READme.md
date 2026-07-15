# Wild Workouts

**Wild Workouts** is a training-session booking platform for personal trainers and their clients — trainers publish
open training slots, clients book them. This repo uses that domain as a hands-on exploration of **Domain-Driven
Design** and **microservice architecture** in Go, adapted from
[ThreeDotsLabs' Wild Workouts example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example). See their
[write-up on the business logic](https://threedots.tech/post/serverless-cloud-run-firebase-modern-go-application/?utm_source=about-wild-workouts#what-wild-workouts-can-do)
for the domain story this implementation follows.

## Services

| Service    | HTTP port | gRPC port | Depends on              |
|------------|-----------|-----------|--------------------------|
| `trainer`  | 4000      | 4100      | —                        |
| `training` | 4200      | —         | `trainer`, `user` (gRPC) |
| `user`     | 4300      | 4400      | —                        |

Each service lives under `internal/<service>` as its own Go module (see `go.work`), following a DDD-style layout:
`domain/`, `app/` (use cases), `adapters/` (infra), `ports/` (HTTP/gRPC entrypoints). Shared code lives in
`internal/common`.

## Tech stack

- **HTTP**: [Echo](https://echo.labstack.com/), contracts defined in OpenAPI, code generated via [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)
- **gRPC**: inter-service calls, [Protobuf](https://protobuf.dev/)-defined
- **PostgreSQL**: queries generated via [sqlc](https://sqlc.dev/)
- **CI/CD**: GitHub Actions

## Getting started

### Prerequisites

- Docker & Docker Compose

### Setup

1. Copy the example env file for each service:

   ```sh
   cp internal/trainer/.env.example internal/trainer/.env
   cp internal/training/.env.example internal/training/.env
   cp internal/user/.env.example internal/user/.env
   ```

2. Start everything:

   ```sh
   docker compose up
   ```

   This builds and runs `trainer`, `training`, `user`, and a `postgres` instance, with hot-reload via
   [`reflex`](https://github.com/cespare/reflex) — source changes under `internal/**` restart the affected service.

## Test strategy

Each service's tests are split into three layers by Go build tag, from fastest/most-isolated to
slowest/most-realistic:

| Layer           | Build tag     | Where                          | What it exercises                                                                              | Needs Postgres? |
|-----------------|---------------|---------------------------------|--------------------------------------------------------------------------------------------------|-----------------|
| Unit            | _(none)_      | `domain/`                      | Pure domain logic in isolation — no DB, no network.                                              | No              |
| Integration     | `integration` | `adapters/db/`                 | Real repositories/read-models against a real Postgres instance.                                  | Yes             |
| Component       | `component`   | `tests/`                        | The service boots for real (real DB, real HTTP/gRPC servers); only external systems are stubbed — auth, and (for `training`) its gRPC calls to `trainer`/`user`. Tests drive it black-box through its generated API client. | Yes             |


