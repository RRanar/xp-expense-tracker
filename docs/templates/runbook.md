# Runbook: <alert name>

<!-- One runbook per alert. Name the file after the alert, lowercase with dashes. Keep it short enough to follow at 3 a.m. -->

- Alert: `<alert name>` (defined in spec/RFC `<id>`)
- Severity: page | ticket
- Owner: <team or person>
- Last reviewed: <date>

## What this means

One or two sentences on the symptom and the user impact.

## Immediate checks (first 5 minutes)

1. Open the dashboard: <link or name>
2. Check recent deploys and config changes.
3. Check the dependency status: <list>

## Diagnosis

| Observation | Likely cause | Next step |
|---|---|---|
| | | |

## Mitigation

- Fast, safe actions first (roll back, disable a flag, scale, fail over).
- What not to do.

## Escalation

Who to contact next and when.

## After the incident

- Record what happened in the incident log.
- Update this runbook and the alert if it was noisy, late or unclear.
