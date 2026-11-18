---
id: 0001
title: Expense Ledger: gRPC service with transactional outbox on GCP
status: Accepted
date: 2026-11-18
approved_by: owner
approved_date: 2026-11-18
---

# RFC 0001: Expense Ledger: gRPC service with transactional outbox on GCP

## Summary

Rework the existing Go expense tracker (HTTP + SQLite, DDD layers) into a small, deployed, production-style service: a gRPC API (`expense-api`) backed by PostgreSQL, publishing domain events to Kafka through a transactional outbox moved by Debezium, consumed by a `notifier` worker, running on GCP Cloud Run and built with Terraform and GitHub Actions. It is a pet project, deployed. It demonstrates Go, gRPC, PostgreSQL, Kafka and GCP end to end, and makes no claim of production experience. Technology choices are recorded in ADRs 0003 to 0012; this RFC ties them into one design, fixes the order of work, and names the risks.

## Motivation

The base repository has clean layering, a TDD/DDD history and passing domain, use-case and HTTP tests, but it does not fully build: the persistence adapter does not compile, the entity UUID is never stored, migrations are not re-runnable, the clock is not injected, and the repository interface has no context, filters, pagination, update or delete. It is also a single HTTP service with SQLite, so it shows none of: contract-first RPC, a relational store with versioned migrations, reliable messaging, cloud delivery or observability.

If nothing is done, the repo stays a half-working sample. The goal is a small but complete system whose design choices can be explained and defended, with failure behaviour that was actually exercised.

Audit findings (verified by building and testing the base repo on 2026-11-18): the persistence adapter does not compile; the entity UUID is never stored; `CREATE TABLE` is not re-runnable; the use case calls `time.Now()` directly; `CreatedAtFromString` hides parse errors; the repository port lacks context, filters, pagination, update and delete; the directory `expence` does not match package `expense`; `go.mod` marks direct dependencies as indirect; no README or CI; the SQLite driver needs cgo. The repair work is spec 0001, which carries the full list.

## Detailed design

### Product scope (kept small)

Personal expense tracking API, synthetic data only.

- Create, read, update, delete expenses (amount in minor units as `int64`, currency, category, merchant, date, note).
- List with filters and pagination.
- Budgets per category and month; a `BudgetExceeded` event when a new expense crosses the budget.
- Export one month as a server stream.
- A `notifier` reacts to `BudgetExceeded` (logs or posts to a webhook; no real email).
- Stretch, behind an interface and a feature flag: receipt photo to structured expense. Out of scope until everything else works.

### Architecture

```
client (grpcurl / small CLI) --HTTP/2--> Cloud Run: expense-api (gRPC)
                                              |  [identity context: login, refresh, JWKS]
                                              '-- PostgreSQL: expenses, budgets, outbox, identity tables
                                                         | WAL (logical decoding, pgoutput)
                                                Debezium + Outbox Event Router
                                                         |
                                              Kafka topic: expense.events.v1 (key = user_id)
                                                         |
                                                notifier (consumer group)
                                                  '-- failures --> expense.events.dlq
```

Two deployable services only; more would add cost without signal. The `identity` bounded context (ADR 0012) ships inside the `expense-api` binary at first, with its own gRPC service, tables and packages, and may be extracted later into a stand-alone OAuth 2.0 token service. The expense API never imports identity code; it verifies tokens through a `TokenVerifier` port, and a CI dependency check enforces this.

- **expense-api** owns business logic. It writes the state change and an `outbox` row in one database transaction and never talks to Kafka, so a broker outage cannot fail a user request (ADR 0007).
- **Debezium** reads the outbox insert from the WAL and publishes it with the Outbox Event Router. Delivery is at-least-once; consumers are idempotent.
- **notifier** is a consumer group. It dedupes by event id in a `processed_events` table, retries a bounded number of times with backoff, then publishes to the DLQ topic. Offsets are committed only after the effect is durable.

### Layering (hexagonal)

```
/api/proto/expense/v1/   contract, buf lint + breaking in CI
/cmd/expense-api, /cmd/notifier
/internal/domain         entities, money, budget rules (no infra imports)
/internal/app            use cases, Clock and repository ports
/internal/adapters       postgres, kafka, token verification
/internal/identity       own domain, use cases, adapters; own proto package (api/proto/identity/v1)
/internal/transport/grpc
/deploy/                 terraform, docker
```

The existing `domain`, `application/usecase`, `infrastructure` split is kept and renamed toward this layout in spec 0001 and later specs; history is preserved. Dependencies point inward only, checked in CI.

### Contracts

- gRPC API: resource-oriented messages, field masks for update, page tokens, a money type, unary CRUD plus server-streaming `ExportMonth`. Generated code is never hand-edited (ADR 0004).
- Events: Protobuf. Types `ExpenseCreated` and `BudgetExceeded`. Each carries event id, type, schema version, occurrence time, correlation id, aggregate id and minimal payload. Key = `user_id`, so one user's events stay ordered within a partition.
- Outbox columns follow the router convention: `id`, `aggregatetype`, `aggregateid`, `type`, `payload`, `created_at`.

### Delivery order

Work is split into specs, one per milestone. Each is written and approved before any code for it.

| Milestone | Spec (planned) | Outcome |
|---|---|---|
| 1 | 0001 | Repo repaired, CI, proto v1, `buf` |
| 2 | later | gRPC CRUD, Postgres adapter and migrations, interceptors, compose |
| 3 | later | Budgets, outbox, Debezium (local), notifier with idempotency, retry, DLQ |
| 3b | later | Failure drills written up in `docs/failure-modes.md` |
| 4 | later | Identity context (JWT, refresh rotation, JWKS), auth interceptor, rate limit, streaming export, error details |
| 5 | later | Terraform, container image, CI/CD with workload identity, Cloud Run deploy |
| 6 | later | Tracing, alerts, load test, README with limitations |

If time is short, drop the streaming export and any GKE variant, never the outbox.

### Decisions recorded as ADRs (Proposed, awaiting the owner)

- ADR 0009: event payload is Protobuf bytes in the outbox (fallback to JSON if the router setup fails the local drill).
- ADR 0010: Kafka is Compose locally and a single self-hosted KRaft node on a small VM for the demo; managed options re-checked on price before milestone 5.
- ADR 0011: Debezium Server as one container beside the broker.
- ADR 0012: self-issued JWT access tokens with rotating refresh tokens, identity as an extractable bounded context.

### Decisions settled inside the spec that needs them

Each is written down in the spec for its milestone and must be settled before that spec is approved, so none is an open RFC question: Kafka client library (milestone 3), whether the `notifier` runs on Cloud Run with min instances 1 and always-allocated CPU, a worker pool if available, or GKE Autopilot (a scale-to-zero consumer is wrong; milestone 5), and the migration tool, goose or golang-migrate, run as a separate job and not at app start (milestone 2).

## Security

Assets: expense and budget records (synthetic only), user identity tokens, database credentials, the replication role.

Data classification:

| Data | Class | Rule |
|---|---|---|
| Amount, currency, category, merchant, date, note | Financial (synthetic only) | Never in logs, metric labels or traces; may be in the database and in event payloads |
| `user_id`, token claims | Internal | `user_id` allowed in logs; token contents never logged |
| Event payload | Financial (synthetic) | Carries ids and the minimum fields the consumer needs; no note text |
| Database and replication credentials | Secret | Secret Manager only |

Token policy (ADR 0012): the project issues its own short-lived JWT access tokens (asymmetric signature, `kid`, JWKS) and opaque refresh tokens that are hashed at rest, rotated on every use, with reuse detection revoking the whole family. Identity is a separate bounded context with its own gRPC service, so it can later be extracted as a stand-alone OAuth 2.0 token service; the expense API only verifies tokens through a port. Exact lifetimes, algorithm and the login flow are fixed in the milestone 4 spec.

Trust boundaries: client to API; API to database; database to Debezium to Kafka; Kafka to notifier; CI to GCP.

| Threat (STRIDE) | Abuse case | Mitigation |
|---|---|---|
| Spoofing | Forged or expired token; stolen refresh token; credential stuffing | Pinned-algorithm JWT validation in a gRPC interceptor (issuer, audience, expiry); refresh rotation with reuse detection; login and refresh rate limits; unauthenticated calls return `Unauthenticated` |
| Tampering | Edit another user's expense | Per-user isolation enforced in the repository layer on every query, tested with two users |
| Repudiation | Disputed change | Structured audit log with request id and user id |
| Information disclosure | Signing key or refresh token leak; secrets in repo or logs; a client or consumer reading the broker without authorisation; one user reads another's events | Secret Manager, `.env.example` only; no amounts or notes in logs; messages keyed and filtered by `user_id` |
| Denial of service | Request flood, huge export | Per-user rate limit, deadlines, page size caps |
| Elevation of privilege | Over-broad cloud identity | One service account per service, least privilege; workload identity federation, no long-lived keys |

The broker requires TLS and authenticated clients when deployed (ADR 0010). The replication role is dedicated with only the privileges Debezium needs, its credentials are in Secret Manager and the database connection uses TLS (ADR 0011). Reflection is enabled only outside the deployed environment. Synthetic data only; no credentials or project ids in the repository.

## Performance

No production load exists. The service is sized for demonstration. The budget is a starting target, tuned by the owner after measurement, not a promise:

- Benchmark tool: `ghz` against the deployed gRPC API, run in milestone 6; `testing` benchmarks for hot domain code where useful.
- Workload model (placeholder, owner tunes): a few synthetic users, a read-heavy mix of list and get with a smaller share of create and update, a handful of concurrent clients. Real values are fixed in the milestone 2 spec.
- Unary CRUD: default budget of p95 under 200 ms and p99 under 500 ms at the model's load, measured; a default, not a claim.
- Event freshness (commit to notifier effect): default p95 under 10 seconds on the demo broker, measured in the milestone 3b drills. The 10 seconds is the target to measure; the 30 seconds in the SLO below is the contract.
- Results are recorded in `docs/benchmarks/` with the exact environment. No numbers are published before they are measured.

## Observability and alerting

### SLIs and SLOs

Starting targets for a pet project, owner-tunable (see the alerting standard). Window 30 days.

| SLI | Good event | SLO (default) |
|---|---|---|
| API availability | gRPC request without a server-side error status | 99.5% |
| API latency | Unary request under 500 ms (default, tunable in the milestone 2 spec) | 95% |
| Event freshness | `BudgetExceeded` effect visible within 30 seconds of commit (default, tunable in the milestone 3 spec) | 99% |

Replication slot lag, DLQ size and consumer lag are cause signals: they are dashboard panels and ticket alerts, not SLIs.

### Alerts

| Alert | Condition (symptom) | Severity | Owner | Runbook |
|---|---|---|---|---|
| AuthAbuseSuspected | Spike of failed logins or refresh-token reuse detections | ticket | owner | [runbooks/auth-abuse-suspected.md](../runbooks/auth-abuse-suspected.md) |
| ApiErrorBudgetBurn | Multi-window burn of the availability SLO (14.4 over 1 h and 5 min as default) | page | owner | [runbooks/api-error-budget-burn.md](../runbooks/api-error-budget-burn.md) |
| EventFreshnessBreach | Events not delivered within the freshness threshold | page | owner | [runbooks/event-freshness-breach.md](../runbooks/event-freshness-breach.md) |
| ReplicationSlotLagHigh | Slot retains growing WAL; connector stalled | ticket | owner | [runbooks/replication-slot-lag.md](../runbooks/replication-slot-lag.md) |
| DatabaseStorageNearFull | Database storage projected to fill within the owner-set horizon, typically from a stalled slot | page | owner | [runbooks/database-storage-near-full.md](../runbooks/database-storage-near-full.md) |
| SigningKeyExpiring | Signing key nears its end of life without a successor in the JWKS | ticket | owner | [runbooks/signing-key-expiring.md](../runbooks/signing-key-expiring.md) |
| BrokerDown | Broker unreachable; cause signal that threatens event freshness | ticket | owner | [runbooks/broker-down.md](../runbooks/broker-down.md) |
| BrokerDiskNearFull | Broker disk usage projected to fill within the owner-set horizon | ticket | owner | [runbooks/broker-disk-near-full.md](../runbooks/broker-disk-near-full.md) |
| DlqGrowing | DLQ message count increasing | ticket | owner | [runbooks/dlq-growing.md](../runbooks/dlq-growing.md) |
| BillingBudgetThreshold | Spend crosses the configured budget | ticket | owner | [runbooks/billing-budget.md](../runbooks/billing-budget.md) |

Alert rules and runbooks are delivered by the specs that build each component and are exercised (fired once) before that spec is Implemented. The runbook files do not exist yet: they are planned stubs, written by the spec that builds each component (replication slot, storage, broker and DLQ by milestone 3, signing key by milestone 4, API and freshness by milestones 2 and 3, billing by milestone 5), and none is relied on until it exists. For a single-owner pet project, "page" means a notification the owner may mute at night.

### Logging and tracing

JSON logs with `request_id`, `user_id`, method, status, duration and event id on the async path. Correlation id travels from the request into the outbox row, the Kafka headers and the notifier logs. Never logged: amounts, notes, tokens, token hashes, passwords, credentials. OpenTelemetry traces cover the gRPC handler, database calls and the notifier handler; a trace link across Kafka is a stretch goal.

## Alternatives considered

- Keep HTTP/REST: rejected, the point is to demonstrate contract-first gRPC (ADR 0004).
- Polling relay instead of Debezium: simpler to run, more code to own; kept as the fallback if Debezium hosting proves too costly (ADR 0007).
- Dual write to database and Kafka: rejected, unsafe.
- Keep SQLite: rejected, needs cgo and does not fit the cloud and CDC goals (ADR 0005).
- More services (separate budget service, gateway): rejected, cost and noise without signal.
- Event sourcing as primary store: larger than the project needs.

## Rollout

Three verification gates must pass before the work they guard: the Debezium binary pass-through drill (ADRs 0009 and 0011, milestone 3), the Cloud SQL logical-decoding documentation check (ADR 0011, before milestone 5) and the check of the token practices against the current RFC 9700, 9068 and 8725 (ADR 0012, before milestone 4). A failed gate supersedes the ADR it guards.

Incremental, one spec at a time, each reviewed and approved before code. Local docker compose is the first environment; the outbox flow is verified end to end locally before anything is created on GCP. Cloud work happens in a throwaway project with a billing budget and alert created first, and a documented teardown. Rollback for the deployed service is redeploying the previous image; infrastructure changes are applied and destroyed by Terraform. Success is each milestone's done-condition from the table above, plus the failure drills in `docs/failure-modes.md`.

## Open questions

None for this RFC. The three that were open are now Proposed ADRs 0009, 0010 and 0011, and the auth decision is Proposed ADR 0012. Two others are closed here:

- The SQLite adapter and the HTTP transport were learning scaffolding and are removed (owner decision, 2026-11-18); spec 0001 carries the removal.
- The receipt-to-expense stretch is out of scope of this RFC. If the owner wants it later, it gets its own RFC.

This RFC should only be accepted together with, or after, review of ADRs 0009 to 0012; they remain Proposed until the owner approves them.
