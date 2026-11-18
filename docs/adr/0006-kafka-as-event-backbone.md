---
id: 0006
title: Kafka as the event backbone
status: Proposed
date: 2026-11-18
supersedes: none
---

# ADR 0006: Kafka as the event backbone

## Context

Domain events (for example an expense created or a budget exceeded) must reach other consumers reliably and in order per aggregate. The owner already works with Kafka and RabbitMQ in production PHP systems, and Kafka is common in the target roles.

## Decision

We will publish domain events to Kafka topics, keyed by aggregate or user id so that related events are ordered within a partition. Events are Protobuf messages with a version header. Consumers are idempotent and use bounded retries and a dead-letter topic (`pet-project/standards/reliable-messaging`). The broker hosting option (managed service on the chosen cloud, a third-party managed Kafka, or self-hosted) is decided in a separate ADR before deployment, after checking current pricing and limits.

## Alternatives considered

- Cloud-native pub/sub service: lowest operational cost, but does not use the skills the project should demonstrate and offers different ordering semantics.
- RabbitMQ: good for task queues, weaker as a replayable event log.
- No broker (direct synchronous calls): couples services and loses the reliability patterns to demonstrate.

## Consequences

Easier: replay, ordering per key, ecosystem for change data capture. Harder: broker operations or cost, partition design, consumer-group behaviour, schema evolution discipline.

## Operational impact

Consumer lag, oldest-message age, retry and dead-letter rates need alerts with runbooks (`pet-project/standards/alerting`, `pet-project/standards/reliable-messaging`). A broker is billed or resourced continuously: set a budget alert and a teardown procedure before creating it. Failure drills for duplicates, restarts and poison messages are required before any messaging spec is Done.
