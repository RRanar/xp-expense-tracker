---
id: 0004
title: gRPC with Protobuf, contract-first
status: Accepted
date: 2026-11-18
approved_by: owner
approved_date: 2026-11-18
supersedes: none
---

# ADR 0004: gRPC with Protobuf, contract-first

## Context

The service needs a typed, versioned external interface, streaming for exports, and consistent error and deadline semantics. The project aims to demonstrate gRPC in a deployed setting.

## Decision

We will define APIs and event schemas in Protobuf, contract-first, with `buf` lint and `buf breaking` checks in CI. Whether generated code is committed or generated in CI is chosen in the first implementation spec. The wire encoding of events through the outbox is decided in ADR 0009. Services expose gRPC; generated code is never edited by hand. Unary and server-streaming calls are used where they fit; deadlines, status-code mapping and the health-checking protocol are mandatory.

Conventions: resource-oriented messages, field masks for updates, opaque page tokens, and a money type (minor units plus currency). Domain errors map to gRPC status codes with `errdetails`, covered by golden tests. All connections use TLS (terminated by the platform when deployed; verified in the deployment spec).

Versioning: packages carry a major version (`expense.v1`). A breaking change means a new `v2` package served alongside `v1` for a deprecation period, never an in-place change. The identity context has its own package (ADR 0012).

## Alternatives considered

- REST/JSON with OpenAPI: broader tooling and easier for browsers, but weaker typing and no native streaming; less aligned with the learning goal.
- GraphQL: flexible for clients, unnecessary complexity for a service-to-service ledger.
- Connect or gRPC-Web on top of Protobuf: friendlier to browsers, but adds a layer and is not needed while clients are a command-line tool and `grpcurl`.

## Consequences

Easier: strong contracts, generated clients, streaming, uniform error model. Harder: browser access needs a gateway or proxy (not planned initially), debugging needs tooling, and some managed platforms need explicit HTTP/2 settings to verify.

## Operational impact

Authentication and rate limiting live in interceptors (`pet-project/standards/security`). Every RPC emits metrics for the availability and latency SLIs (`pet-project/standards/alerting`). The health-checking service is on everywhere; server reflection is enabled for local development only and is off when deployed. Platform support for gRPC on the chosen runtime is to be verified in the deployment spec.
