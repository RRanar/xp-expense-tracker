---
name: new-adr
description: Record one architecture decision as a numbered ADR in the active target project, from the kit template. Status starts Proposed; only the user accepts it. Use for any decision with trade-offs (database, protocol, pattern, deployment target).
argument-hint: <decision title>
---

# new-adr

One decision per ADR, written into the **target project's** `docs/adr/`, not into a shared kit. Writing the ADR is how a decision gets made here, not paperwork after the fact.

## Steps

1. Resolve the target: `python3 -I tools/docgate.py target` (name and path). If none is registered, send the user to `/init-project`.
2. Read the standards the decision touches (`docs/standards/`) and the target's `docs/adr/` for decisions it may conflict with or supersede.
3. Scaffold: `python3 -I tools/docgate.py new adr "<title>"`. It prints the absolute path of the new file in the target.
4. Fill every section with specifics:
   - **Context**: forces, constraints, requirements. Link the RFC or spec that triggered it.
   - **Decision**: one or two sentences, "We will ...".
   - **Alternatives considered**: at least two real options, each with the reason it lost.
   - **Consequences**: easier, harder, and when to revisit.
   - **Operational impact**: security, observability, alerting and cost consequences; name any alert or runbook it requires.
5. If it replaces an earlier ADR, set `supersedes: <id>` and add a one-line pointer in the old ADR's Consequences (do not rewrite the old decision).
6. `python3 -I tools/docgate.py check`, fix structural problems.
7. Leave status `Proposed`. Tell the user the path and suggest `/review-doc <path>`.

## Rules

- Never set `Accepted` or fill `approved_by` / `approved_date`. Only the user approves, via `/review-doc`.
- Technology names belong in ADRs, never in the kit's `docs/standards/`.
- Do not invent facts about performance, cost or platform limits. Write "to verify" and say how.
- Do not copy another project's decisions; derive them from this project's context and the user's choices.
