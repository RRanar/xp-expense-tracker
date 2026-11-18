---
id: 0009
title: Event payload format for the outbox: Protobuf bytes
status: Proposed
date: 2026-11-18
supersedes: none
---

# ADR 0009: Event payload format for the outbox: Protobuf bytes

## Context

The Outbox Event Router needs a payload column. RFC 0001 defines events as Protobuf messages and the API contract is Protobuf with `buf` (ADR 0004). The router can carry a JSON payload most simply, or raw bytes with a byte-array converter. A web check on 2026-11-18 found that Debezium documents `io.debezium.converters.BinaryDataConverter` (and, in some versions of the docs, Kafka's `ByteBufferConverter`) for passing a binary payload through the router unchanged, that one issue reports a `ByteArrayConverter` base64-encoding the payload, and that the docs disagree with each other; support on Debezium Server is requested in DBZ-7512 and its current status is unconfirmed. The exact setup therefore must still be verified against the Debezium version in use.

## Decision

We will store the event as Protobuf-encoded bytes in a `bytea` outbox payload column (produced by `proto.Marshal` in the outbox write, inside the same transaction as the state change) and publish it unchanged with a Debezium binary pass-through converter (exact converter name verified against the Debezium version in use), so events share the schema workflow and compatibility checks (`buf breaking`) of the API. Event id, type, schema version and correlation id travel as Kafka headers; the milestone 3 spec must confirm the router can emit all four, and otherwise carries the missing ones in the outbox payload envelope. A named drill in the milestone 3 spec proves that the bytes arrive at the consumer unchanged and decode to the original message through the chosen Debezium host. If that drill fails, we choose one fallback and supersede this ADR (and ADR 0011 if the host changes): a JSON payload on Debezium Server, Kafka Connect as the host, or the polling relay of ADR 0007; the Protobuf schema remains the source of truth either way.

## Alternatives considered

- JSON payload: simplest with the router, human-readable in the broker, but a second schema definition to keep in sync and no compile-time contract.
- JSON text wrapping base64 Protobuf: works anywhere but adds size and a decode step for no gain.
- Schema registry with Avro: heavier infrastructure than the project needs.

## Consequences

Easier: one contract language, generated types in producer and consumer, compatibility checked in CI. Harder: messages are not readable with plain console tools (use a decode helper), and the router converter configuration must be proven early.

## Operational impact

The pass-through drill is the gate for this ADR: until it passes, treat this decision as unproven.

Incompatible schema changes are caught by `buf breaking` in CI; breaking changes need a new ADR and a version bump (`reliable-messaging` standard). The DLQ keeps the original bytes and headers so a message can be replayed or decoded later. Add a decode helper to the DLQ runbook.
