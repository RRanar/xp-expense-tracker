---
id: 0003
title: Use Go for the expense ledger
status: Accepted
date: 2026-11-18
approved_by: owner
approved_date: 2026-11-18
supersedes: none
---

# ADR 0003: Use Go for the expense ledger

## Context

The project exists to demonstrate a contract-first gRPC service with a relational store, messaging and cloud delivery. A language is needed with first-class gRPC support, small static binaries and a strong concurrency model. An existing Go repository (an expense tracker with a layered, test-driven structure) can be reworked instead of starting over.

## Decision

We will implement the ledger services in Go, evolving the existing expense tracker repository. Builds use `CGO_ENABLED=0` and pure-Go drivers only; a dependency that needs cgo is not accepted without a new ADR. The domain stays free of framework imports (see `pet-project/standards/architecture-ddd`).

## Alternatives considered

- PHP/Symfony: mature and productive for request-response web work, but gRPC, long-lived consumers and static single-binary deployment are not its natural fit.
- Node.js/TypeScript: good gRPC libraries and fast iteration, but a runtime must be shipped with the service and typing of generated contracts is less strict.
- Java/Kotlin: strong ecosystem, larger setup cost and no head start from existing code.

## Consequences

Easier: single static binaries, small container images, strong concurrency and tooling, a direct signal for Go roles. Harder: the owner is less fluent in Go, so progress will be slower at first and idioms must be learned deliberately. This is a pet project, so it demonstrates learning and design, not years of Go experience.

## Operational impact

Static binaries without a C toolchain (pure-Go drivers) keep images minimal and builds reproducible. The Go toolchain version is pinned in `go.mod` (currently `go 1.25.3`) and upgraded deliberately, not implicitly. Dependency and vulnerability scanning runs in CI with `govulncheck` (spec 0001; `pet-project/standards/security`). Exact tool and library choices are made in later ADRs or specs and to be verified when selected.
