# Standard: reliable messaging and events

Platform-agnostic. Applies to any asynchronous communication between services.

## Facts to design around

- Delivery is **at-least-once** in practice. Duplicates will happen. "Exactly once" holds only inside a system that controls both ends; design for duplicates instead.
- Ordering is guaranteed only within a defined partition of the stream, if at all.
- Consumers crash, restart, rebalance, and fall behind.
- A write to a database and a publish to a broker cannot be made atomic by trying harder. Do not dual-write.

## Patterns

- **Transactional outbox**: the service writes its state change and an outbox record in the same local transaction. A separate mechanism moves outbox records to the broker (log-based change data capture, or a polling relay). The business request never depends on broker availability.
- **Idempotent consumer**: every handler tolerates duplicates. Record processed event ids with the effect in one transaction, or make the operation naturally idempotent (upsert by key, set rather than increment).
- **Ordering by key**: choose the partition key so that events needing order share a key (for example all events of one aggregate). Document the key in the event contract.
- **Retry with backoff and jitter** for transient failures, bounded in count and time.
- **Dead-letter path** for messages that keep failing: park them with the error and original headers, alert on growth, and provide a documented replay procedure.
- **Poison-message isolation**: one bad message must not block the stream.
- **Commit after success**: advance the consumer position only after the effect is durable.

## Event contracts

- Events are facts in the past tense, immutable, and owned by the producing context.
- Each event has: a unique id, type, schema version, occurrence time, a correlation id, the aggregate id, and a minimal payload. No secrets or unneeded personal data.
- Schemas are explicit, versioned, and checked for backward compatibility in CI. Breaking changes need an ADR and a migration plan (dual publish, version bump).
- Consumers ignore unknown fields (tolerant reader).
- Prefer events that carry enough data to avoid a synchronous callback, without copying the whole aggregate.

## Operating the pipeline

- Monitor: consumer lag and oldest-message age, processing latency, retry and dead-letter rates, outbox backlog age, and the health of whatever moves outbox records (including any replication position that can accumulate storage when stalled).
- Define per stream: retention, partition count rationale, replication/durability settings, and the replay procedure.
- Alert on user-visible symptoms (events not delivered within the freshness SLO) and ticket on early signs (growing lag, dead-letter growth). See `alerting.md`.

## Failure drills (required before Done)

1. Broker down while the service keeps accepting writes: no loss, events flow after recovery.
2. The outbox mover is stopped and restarted: no loss, duplicates tolerated.
3. Consumer killed mid-processing and restarted: no lost work, no double effect.
4. The same message delivered twice: one effect.
5. A poison message: lands in the dead-letter path, stream continues, alert fires.
6. Replay from the dead-letter path: succeeds and is idempotent.

## Review checklist

- [ ] Any dual write? Replace with the outbox.
- [ ] Is every consumer idempotent, and how is that tested?
- [ ] Is the ordering key chosen and documented?
- [ ] Are retries bounded, with a dead-letter path and replay procedure?
- [ ] Are event schemas versioned and compatibility-checked?
- [ ] Are lag, backlog age and dead-letter alerts defined with runbooks?
- [ ] Were the failure drills run?
