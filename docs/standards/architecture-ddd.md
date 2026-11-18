# Standard: architecture and DDD

Platform-agnostic. Applies to any backend service in this project.

## Strategic design

- **Ubiquitous language.** One glossary per bounded context (`docs/specs` link to it). Code, tests, docs and conversations use the same words. A new word in code means a glossary entry.
- **Bounded contexts.** A model is valid inside one context. Name the contexts, their owners, and the relationships between them (upstream/downstream, customer/supplier, conformist, anti-corruption layer). Draw a context map once there is more than one.
- **Do not share domain models across contexts.** Share contracts (events, APIs), not entities.
- **Anti-corruption layer** at every boundary to a model we do not control: translate, do not leak.

## Layers and the dependency rule

```
presentation / transport  ->  application  ->  domain  <-  infrastructure
```

- **Domain**: entities, value objects, aggregates, domain services, domain events, repository interfaces (ports). Imports nothing from other layers or frameworks.
- **Application**: use cases (one per command or query), transaction boundaries, orchestration. Depends on the domain only.
- **Infrastructure**: adapters implementing ports (persistence, messaging, clock, identity). Depends inward.
- **Presentation / transport**: translates external requests to use-case input and results to responses. No business rules.
- Dependencies point inward. **The dependency rule is checked in CI** (an architecture test or linter); a violation fails the build.
- The composition root is the only place that wires concrete adapters.

## Tactical building blocks

- **Entity**: identity and lifecycle; compared by id.
- **Value object**: immutable, compared by value, validated at construction (money, currency, time instants, identifiers, email). Invalid values cannot exist.
- **Aggregate**: a consistency boundary with a root. Invariants hold within one aggregate per transaction. Reference other aggregates by id. Keep aggregates small.
- **Domain service**: a domain operation that does not belong to one entity.
- **Domain event**: a fact that happened, past tense, immutable, versioned. Raised by aggregates, published after the transaction commits (see `reliable-messaging.md`).
- **Repository**: a port for loading and saving aggregates; speaks in domain terms, not storage terms.
- **Factory**: when construction is complex or must enforce invariants.

## Rules of thumb

- **Make illegal states unrepresentable.** Validate in constructors; no public setters; no anemic objects holding rules elsewhere.
- **Money**: integer minor units plus a currency; never floating point. Rounding rules are explicit and tested.
- **Time and identifiers are ports.** Inject a clock and an id generator so tests are deterministic.
- **Errors**: a small taxonomy of domain errors (invalid input, not found, conflict, forbidden, unavailable); map them to transport status at the edge only. Do not hide parse failures behind defaults.
- **Every operation takes a cancellation/deadline context** and every outbound call has a timeout.
- **Commands change state, queries do not.** Separate read models only when a measured need appears (record it in an ADR).
- **Persistence ignorance**: the domain does not know how it is stored. Reconstitution from storage uses a dedicated path that does not generate new identities.
- **No premature microservices.** Split a service only for an independent scaling, ownership or release need, recorded in an ADR.

## Contracts

- Contract-first for external interfaces: the contract is a reviewed artefact with automated lint and breaking-change checks.
- Backward-compatible evolution by default; breaking changes need an ADR and a migration plan.
- Event schemas are versioned and documented like APIs.

## Documentation obligations

- A new aggregate, context boundary, or dependency direction needs an ADR.
- Each spec's "Domain model" section names the aggregates, invariants, events and glossary terms it touches.
- Architecture diagrams live in `docs/` as text (diagram-as-code) and change with the code.

## Review checklist

- [ ] Does the domain import anything outside itself?
- [ ] Are invariants enforced in the aggregate rather than in handlers?
- [ ] Are value objects validated at construction?
- [ ] Are time and id generation injected?
- [ ] Is the transaction boundary one aggregate?
- [ ] Are errors mapped to transport status only at the edge?
- [ ] Is the dependency-rule check in CI and green?
