# Project instructions

This project follows a docs-first workflow. The skills, gate tool, standards and templates are local to this repository (copied from the pet-project kit; refresh them with `python3 -I tools/docgate.py update` run from the kit).

## The rule

**No implementation code without an Accepted spec.** The chain is:

`RFC (when risky or cross-cutting) -> ADR(s) -> spec -> user approval -> implement (TDD) -> alerts, runbooks, benchmarks -> spec Implemented`

`tools/docgate.py` enforces it through a PreToolUse hook (`.claude/settings.json`). For an edit inside the guarded code paths (see `.docgate.json`), `docs/work/ACTIVE` must name a spec that is Accepted, whose linked ADRs/RFCs are Accepted, and the file must be inside the spec's `scope`. Docs and files outside the code paths are not blocked.

## Hard rules

- **Only the user approves.** Status moves Proposed -> Accepted only when the user says so in the conversation; then record `approved_by` and `approved_date`. Never accept your own document.
- **Do not work around the gate.** No editing code through Bash (`sed`, `cat >`, `tee`, scripts). If the gate blocks you, fix the documents. File operations a spec explicitly authorises (for example `git mv`, `git rm`) are allowed.
- **Do not change Claude Code permissions** or settings on your own. Tell the user the setting; they apply it.
- **Alerting is part of the work.** Every spec and RFC has an "Observability and alerting" section: SLIs/SLOs, alerts with severity and owner, a runbook per alert (or a reasoned "not applicable"). A feature is not done until its alert rules and runbooks exist and were exercised.
- **TDD.** Failing test first, from an acceptance criterion; then minimal code; then refactor.
- **Never invent.** No fabricated benchmark numbers, incident history or production claims. This is a pet project: say "pet project, deployed", never "production experience".
- **No secrets, no real data.** Synthetic data only.
- **Code explains itself; comments are the exception.** The best comment is clear code: precise names, small functions, types that make invalid states unrepresentable. Do not write comments that restate what the code does, mark sections, or keep dead code. Write a comment only when it adds what the code cannot say (a non-obvious reason, a constraint, a trade-off), and always write the module or package-level documentation comment that the language's convention expects (for example a Go package comment) describing what the module is for.
- Commits reference the spec id (`spec-0001: ...`). Do not commit or push unless the user asks.

## Skills (in `.claude/skills/`)

`/new-rfc`, `/new-adr`, `/new-spec`, `/review-doc`, `/implement`, `/alert-design`.

Helper: `python3 -I tools/docgate.py status|check|new|target`.

## Where things live

`docs/adr/`, `docs/rfc/`, `docs/specs/`, `docs/runbooks/`, `docs/alerts/`, `docs/benchmarks/`, `docs/work/ACTIVE`; process in `docs/process/workflow.md`, standards in `docs/standards/` (platform-agnostic defaults the owner tunes), templates in `docs/templates/`.
