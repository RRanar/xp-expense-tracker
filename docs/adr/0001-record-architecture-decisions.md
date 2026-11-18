---
id: 0001
title: Record architecture decisions
status: Proposed
date: 2026-11-18
supersedes: none
---

# ADR 0001: Record architecture decisions

## Context

Decisions made in conversation or in code are forgotten and re-argued. The reasoning behind choices is part of what this project should demonstrate.

## Decision

We will record each significant architecture decision as a numbered ADR in `docs/adr/`, written from the pet-project ADR template, with a status. Status moves from Proposed to Accepted only by the owner's explicit approval. Accepted ADRs are not rewritten; a later ADR supersedes them.

## Alternatives considered

- Decisions in commit messages or the README: hard to find and no alternatives recorded.
- An external wiki: drifts from the code and cannot be read by the docs-first gate.
- No record: cheapest now, expensive at every later "why did we do this".

## Consequences

Easier: onboarding, reviews, interviews, reversing a decision knowingly. Harder: a small overhead per decision and the discipline of writing the ADR before the code.

## Operational impact

None at runtime. The gate reads ADR status to decide whether implementation may start.
