---
id: 0002
title: Docs-first workflow with an automated gate
status: Proposed
date: 2026-11-18
supersedes: none
---

# ADR 0002: Docs-first workflow with an automated gate

## Context

Rules written only in documents are skipped under time pressure, especially when an AI assistant writes much of the code. The project should practise designing security, testing, benchmarking, logging and alerting before implementation.

## Decision

We will follow the pet-project workflow: RFC when risky or cross-cutting, then ADRs, then a spec, then the owner's approval, then test-first implementation, then alert rules, runbooks and benchmarks. The pet-project gate (a Claude Code PreToolUse hook) blocks edits to this project's code paths unless `docs/work/ACTIVE` names an Accepted spec whose linked ADRs/RFCs are Accepted and the file is inside the spec's scope. Documentation is always editable. Alerting design is a mandatory section of every spec and RFC.

## Alternatives considered

- Guidance in a CLAUDE.md only: advisory and easily bypassed.
- CI-only checks on pull requests: catches violations late, after code exists.
- A heavyweight process tool: unnecessary for one person.

## Consequences

Easier: consistent process, early thought about security and alerting, a clear audit trail. Harder: friction for small changes (small specs and the docs-only exemption mitigate it). Limits: the hook sees Write, Edit and MultiEdit only, not shell-driven edits, and cannot verify who approved a document; explicit rules cover both.

## Operational impact

No runtime impact. A failure of the gate on a code path blocks the edit (fail closed); on other paths it never blocks. Standards used by the workflow (security, testing, logging, alerting, benchmarking, architecture, reliable messaging) live in pet-project and are referenced by name from specs.
