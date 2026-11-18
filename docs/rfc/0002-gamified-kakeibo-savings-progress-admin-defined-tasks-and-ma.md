---
id: 0002
title: Gamified Kakeibo savings: progress, admin-defined tasks and manual check-offs
status: Proposed
date: 2026-11-19
---

# RFC 0002: Gamified Kakeibo savings: progress, admin-defined tasks and manual check-offs

## Summary

Add a **Progress** bounded context to the expense ledger. It turns what a user does (logging expenses, staying inside a budget, completing a month-end reflection, ticking off a manual task) into XP, levels and completed tasks, following the habits of **Kakeibo**, the Japanese household-ledger method for saving. Tasks are data defined by an admin, not code: the admin panel is built later, so this RFC fixes the task model and the admin API boundary now and seeds tasks by migration until the panel exists. Progress is computed from the domain events of RFC 0001 by a separate consumer, so the ledger stays small and unaware of gamification.

This is a **hypothesis to test**, not a known result: that Kakeibo-style prompts plus visible progress make a user record and review spending more consistently. Nothing here claims the effect exists. Section "Rollout" says how it will be checked, and what a pet project with one real user can and cannot show.

## Motivation

Problem: a ledger only records. A user who forgets to log, or never looks back, gets no value from it. Kakeibo's value is the habit, not the table: record spending, group it, and reflect at the end of the month. RFC 0001's `BudgetExceeded` event covers the one moment a user overspends. It does nothing for the many days when they simply keep up.

Who is affected: the owner as the first user; later any user of the service.

Cost of doing nothing: the project stays a CRUD ledger with one alert-style event. That is acceptable (it is already a complete RFC 0001 deliverable); this RFC is optional scope on top, and the design below keeps it removable.

### Kakeibo, as far as this RFC relies on it

Facts below were checked against secondary sources, not the book. Treat the details as editable assumptions.

- Origin: credited to Motoko Hani, first published in a women's magazine in 1904 ([Wikipedia: Kakeibo](https://en.wikipedia.org/wiki/Kakeibo)). The English-language popularisation is Fumiko Chiba's 2018 book.
- Spending is grouped in four buckets, commonly Survival, Optional, Culture and Extra. Editions differ: some fold Culture into Optional, and Wikipedia names them Essentials, Non-essential needs, Cultural and Non-anticipated ([Money Crashers](https://www.moneycrashers.com/kakeibo-japanese-budgeting-method-save-money/), [PocketGuard](https://pocketguard.com/blog/kakeibo-method/)). **The bucket names and count are therefore configuration, not a hard-coded fact.**
- A monthly routine: write down income and fixed costs at the start, record purchases through the month, then reflect with four questions: how much money is available, how much to save, how much is being spent, how to improve ([Fatherly](https://www.fatherly.com/love-money/kakeibo-financial-approach-budgeting/) confirms two of them directly; the other two come from search-result summaries of Chiba's book and are not verified against the book itself).
- It is traditionally handwritten. We keep the reflection as free text typed by the user; no automatic scoring of its content.

## Detailed design

### Contexts

| Context | Owns | Notes |
|---|---|---|
| Ledger (RFC 0001) | expenses, budgets, outbox | Unchanged except small additions below |
| **Progress (new)** | task definitions, assignments, progress, XP awards, reflections | Separate schema, separate consumer group |
| Identity (ADR 0012) | users, tokens, roles | Supplies the `admin` role |

Progress does not import ledger code. It learns what happened only through events, the same contract style as the `notifier`.

### Kakeibo concepts added to the ledger

Small, and only because tasks need them:

- **Bucket**: each existing user category maps to one of the configured Kakeibo buckets. Expenses keep their free-form category; the mapping is a separate table (user-owned, with sensible defaults). A new `bucket` field on the `ExpenseCreated` payload is added by the ledger at event time, so Progress never queries the ledger.
- **Month plan**: income, fixed costs and a savings goal for a month (minor units plus currency, as everywhere). Needed for "reached the savings goal" style tasks. Optional for a user.
- **Reflection**: for a closed month, the user writes the four answers as text. Stored in Progress, never logged, never used in a rule beyond "was submitted".

### Progress model

```
TaskDefinition (versioned, immutable once published)
  id, version, title, description, kind, params, reward_xp, period, active_from, active_to
  kind: AUTOMATIC rule | MANUAL check-off
  automatic rules (a fixed set, each with typed params):
    LOG_COUNT        log N expenses in a period
    LOG_STREAK       log on N consecutive days
    BUDGET_ADHERENCE finish a month at or under a category budget
    SAVINGS_GOAL     month savings reach the plan's goal
    REFLECTION_DONE  submit the monthly reflection
Assignment   one user + one task definition version + one period
Progress     counters and state per assignment (OPEN, DONE), derived from events
XP ledger    append-only awards: user, assignment, xp, reason, source event id
Level        derived from total XP by a pure function (thresholds are config)
```

Kanban-style views are not part of this; the user's "board" is a read model over assignments.

Rules are a **closed set of typed rules with parameters**, not a scripting or expression language. An admin chooses a rule and fills its parameters. New rule types need code and a spec. This keeps the admin panel safe to build later (no admin-authored logic runs on the server).

### Data flow

```
 expense-api ──ExpenseCreated/Updated/Deleted, BudgetExceeded, MonthClosed──► outbox ► Debezium ► Kafka
                                                                                                │
                                                       progress consumer group (key = user_id) ◄┘
                                                         │ dedupe by event id (processed_events)
                                                         │ evaluate rules for the user's open assignments
                                                         ▼
                                              PostgreSQL: progress, xp_ledger, assignments
                                                         ▲
 client ── ProgressService (gRPC) ── board, XP, level, manual check-off, submit reflection
 admin  ── AdminTaskService (gRPC, role admin) ── create/publish/retire task definitions  [future panel]
```

- Events are consumed idempotently: a redelivered event has one effect (same pattern as the `notifier`, RFC 0001). Offsets are committed after the XP and progress rows are durable.
- **New events** the ledger must emit beyond RFC 0001's two: `ExpenseUpdated`, `ExpenseDeleted` and `MonthClosed`. See open questions for who produces `MonthClosed`.
- **Awards are never revoked.** If an expense is later deleted, history stays; this avoids rewriting XP, and ledger edits cannot be used to farm or erase progress silently. The cost is that a user can edit data after earning. Accepted for a pet project; recorded in the ADR.
- **Manual check-offs** are self-attested, therefore capped: a per-period XP ceiling for manual tasks, and a task can be ticked once per assignment. No automatic task can be completed by hand.
- **Assignment**: at the start of each period (a scheduled job, one per period), active task definitions are assigned to users. A definition edit creates a new version; open assignments keep the version they started with.

### Admin boundary

- The admin panel is a future client. Until then definitions are seeded by migration or a small CLI run by the owner.
- The API is already shaped for it: `AdminTaskService` (create draft, publish, retire, list) behind the `admin` role from the identity context. Draft, then publish, with an audit row for every admin action (who, when, what changed).
- Published definitions are immutable. Corrections are new versions.
- The role claim format and where roles are stored are decided in the identity spec (milestone 4); this RFC only requires the claim to exist.

### Contracts

- `api/proto/progress/v1/progress.proto` (client) and `api/proto/progress/v1/admin.proto` (admin). Contract-first, `buf` lint and breaking checks, as in spec 0001.
- Events: new Protobuf messages in the existing event package, versioned like the others (ADR 0009). Payloads carry ids, bucket and period only: **no amounts, notes or merchants** beyond what a rule needs (a rule that needs an amount, such as `SAVINGS_GOAL`, uses the minor-unit sum computed in the ledger and sent as a field of `MonthClosed`).

### Delivery order (proposal)

Depends on RFC 0001 milestone 3 (outbox, Debezium, notifier pattern). Nothing here starts before that exists.

| Step | Outcome |
|---|---|
| P1 | Spec: Kakeibo concepts in the ledger (bucket mapping, month plan) and the new events |
| P2 | Spec: Progress core: definitions, assignments, XP ledger, `LOG_COUNT`, `LOG_STREAK`, manual check-off, idempotent consumer, DLQ |
| P3 | Spec: `BUDGET_ADHERENCE`, `SAVINGS_GOAL`, reflection, month close |
| P4 | Spec: admin API and audit (panel itself is a separate project) |

## Security

Assets: progress and XP records (low value, but they reveal activity and goals), reflections (free text that can contain sensitive personal finance details), task definitions (integrity matters, they drive awards), the admin role.

Trust boundaries: client to ProgressService; admin client to AdminTaskService; Kafka to the progress consumer.

| Threat (STRIDE) | Abuse case | Mitigation |
|---|---|---|
| Spoofing | A caller poses as another user to read or tick tasks | Identity from the verified token only; every query scoped to the token's user; another user's record is "not found" (as in spec 0001) |
| Spoofing | A caller poses as admin | Admin API requires the `admin` role from a verified token; separate service and proto, so it can be disabled or network-restricted independently |
| Tampering | Forge events to award XP | Only the outbox path produces events; Kafka produce rights limited to the ledger's identity; consumer validates schema version and ignores unknown types |
| Tampering | Edit a published task to retroactively change awards | Published definitions immutable; changes are new versions |
| Tampering | Farm XP with manual check-offs | Manual tasks capped per period, one tick per assignment, lower reward by policy |
| Repudiation | Admin changes a task and denies it | Audit row per admin action, append-only |
| Information disclosure | Reflections or amounts leak through logs or events | Reflections never logged and not placed in events; events minimal; logging standard applies (no amounts, notes) |
| Information disclosure | Leaderboards expose other users' behaviour | **No leaderboard or cross-user view.** Each user sees only their own progress |
| Denial of service | Mass assignment job or a hot user floods the consumer | Per-user ordering by key; bounded retries then DLQ; assignment job batched; rate limit on check-off and reflection endpoints |
| Elevation of privilege | Admin-authored rule executes code | Closed set of typed rules, no expression language |

Data classification: progress records internal; reflections financial-personal (synthetic only in this project); definitions internal. Synthetic data only, as always.

## Performance

No load exists and none is claimed. Workload model (placeholder, the owner tunes): a few synthetic users, tens of events per user per day, one assignment run per period.

Starting targets, to be **measured**, not promised:

- Event freshness (ledger commit to visible progress): default p95 under 10 seconds on the demo broker, the same default RFC 0001 uses for the notifier path.
- Board read (`GetBoard`): default p95 under 200 ms, p99 under 500 ms, the same defaults as unary CRUD.

Benchmark plan: `ghz` for the read and write RPCs, and an event replay test that feeds N synthetic events and measures time to a consistent XP total. Real numbers go to `docs/benchmarks/` when the specs that build this are implemented.

## Observability and alerting

### SLIs and SLOs

| SLI | Good event | SLO (default, tunable in the P2 spec) | Window |
|---|---|---|---|
| Freshness | Award or progress visible within 30 s of the ledger commit | 99% | 30 days |
| Correctness | A reconciliation job recomputing XP from the event history matches the stored total | 100% of checked users (any mismatch is a defect) | continuous |
| Availability | ProgressService unary request without a server-side error | 99.5% | 30 days |

Admin API availability has no SLO: it is a low-volume owner tool.

### Alerts

Runbook files do not exist yet. They are planned stubs, written by the spec that builds each component and exercised before that spec is Implemented. None is relied on until it exists.

| Alert | Condition (symptom) | Severity | Owner | Runbook |
|---|---|---|---|---|
| ProgressFreshnessBurn | Multi-window burn of the freshness SLO (14.4 over 1 h and 5 min as default) | page | owner | [runbooks/progress-freshness-burn.md](../runbooks/progress-freshness-burn.md) |
| ProgressConsumerAbsent | No events processed for a period while the ledger is emitting (dead-man's switch) | page | owner | [runbooks/progress-consumer-absent.md](../runbooks/progress-consumer-absent.md) |
| ProgressDlqNotEmpty | Poison event in the progress DLQ topic | ticket | owner | [runbooks/progress-dlq.md](../runbooks/progress-dlq.md) |
| XpReconciliationMismatch | Reconciliation finds a stored XP total that differs from the recomputed one | ticket | owner | [runbooks/xp-reconciliation.md](../runbooks/xp-reconciliation.md) |
| ProgressApiErrorBudgetBurn | Multi-window burn of the availability SLO | page | owner | [runbooks/progress-api-error-budget-burn.md](../runbooks/progress-api-error-budget-burn.md) |

For a one-person pet project "page" means a notification the owner may mute at night.

### Logging and tracing

JSON logs with `request_id`, `user_id`, method, status, duration, and `event_id` on the async path; the correlation id travels from the ledger request through the outbox row and Kafka headers into the consumer. Admin actions log `admin_id`, `task_id`, `task_version`, action. **Never logged:** reflections, amounts, notes, tokens. Traces cover the gRPC handlers, database calls and the consumer handler; a trace link across Kafka remains a stretch goal, as in RFC 0001.

## Alternatives considered

| Option | Verdict |
|---|---|
| **Gamification inside the ledger service** (extra tables and RPCs in `expense-api`) | Simplest to ship and has no new event types for awards. Rejected as the default: it couples a speculative feature to the core ledger and its SLOs; a gamification bug could break expense writes. It is the fallback if the consumer approach costs too much. |
| **Separate consumer and API in the same repo, own schema (chosen)** | Matches the `notifier` pattern, isolates failure, keeps the ledger unaware. Costs a second binary and the new events. |
| **Separate service with its own database** | Cleanest ownership, but RFC 0001 already rejects extra services for cost and noise without signal. Rejected for now; revisit only if an independent scaling or ownership need appears. |
| **Admin-authored rules in an expression language (CEL or similar)** | Maximum flexibility for the future panel. Rejected: runs admin-authored logic on the server, harder to test and secure. A closed set of typed rules covers the stated needs; revisit with an ADR if it does not. |
| **Hard-code tasks, no admin model** | Fastest. Rejected: the stated goal is admin-managed tasks later; retrofitting versioning and immutability is harder than starting with them. |
| **Revocable awards (recompute XP from the current ledger)** | Fairer after edits and deletes. Rejected as default: complex, makes XP non-monotonic and surprising. Recorded as an option in the ADR. |
| **Leaderboards / social comparison** | Common in gamification. Rejected: exposes spending behaviour and is out of scope for a single-owner pet project. |

## Rollout

Gated on RFC 0001 milestone 3 (outbox and Kafka working with idempotent consumers). The whole feature sits behind a flag: with it off, the ledger emits the extra events (harmless) and nothing consumes them. Rollback is turning the consumer off; Progress data is a separate schema and can be dropped without touching ledger tables.

**How the hypothesis is checked, honestly.** A pet project has one real user, so it cannot show a causal effect. What it can show:

1. The mechanism works: synthetic-user simulations feed event streams and the XP totals match the reconciliation job, including duplicates, restarts and a poison event (the failure drills from the testing standard).
2. The owner's own use over a fixed period, written up as observation: days with at least one logged expense and month-end reflections completed, compared with the owner's own earlier period. This is an anecdote, reported as one. No claim of behaviour change beyond that.

If the owner does not find the loop useful after that period, the feature is removed (flag off, schema dropped) and the outcome recorded in an ADR.

## Open questions

Each needs an answer before the spec it blocks can be approved.

| # | Question | Blocks | Owner |
|---|---|---|---|
| 1 | Is "Kobane" Kakeibo (assumed here)? Which Kakeibo elements matter most: the four buckets, the monthly reflection, the savings goal, or all three? | P1 | owner |
| 2 | Bucket names and count: the four-bucket default (Survival, Optional, Culture, Extra), or three with Culture folded into Optional? | P1 | owner |
| 3 | Who produces `MonthClosed`: a scheduler (for example Cloud Scheduler), a user action ("close month"), or the first event of the next month? | P3 | owner |
| 4 | XP and level numbers: reward sizes, level thresholds and the manual-task cap. Placeholder values only until the owner sets them; none is invented here | P2 | owner |
| 5 | Admin role claim and storage: decided in the identity spec (milestone 4). Until then, is a CLI-seeded definition set acceptable for P2 and P3? | P2, P4 | owner |
| 6 | Is the progress consumer a worker in the `notifier` process or its own binary? Own binary assumed here | P2 | owner |
| 7 | Awards never revoked (assumed) versus recomputed on edit | P2 | owner |
| 8 | Where is the monthly reflection text stored: in Progress (assumed, because it is not ledger data) or in the ledger? | P3 | owner |

### ADRs this RFC implies (to be written with `/new-adr`, Proposed)

1. Progress as a separate consumer context with its own schema, not part of the ledger.
2. A closed set of typed task rules, no expression language.
3. Append-only XP ledger; level derived; awards never revoked.
4. Kakeibo bucket mapping as configuration over free-form categories.
5. Admin API boundary: separate service and proto, role-gated, immutable published task versions, audited.
