# Standard: logging

Platform-agnostic. Logs explain what happened to one request or one piece of work. Metrics say how often; traces show where time went. Use each for its job.

## Principles

1. **Structured, machine-readable events** (one JSON object per line), not free text. Messages are constant strings; variable data goes in fields.
2. **Every log line is correlatable**: it carries a correlation/request id and, when tracing is on, the trace and span ids. The id is created at the edge and propagated across calls and messages.
3. **Log at the boundary, once.** Handle an error where you can act on it; log it where it ends. Do not log and rethrow at every layer.
4. **Never log secrets or personal data.** Redact by default; allow-list fields rather than dumping objects.
5. **Logs are for diagnosis, not for metrics or audit.** Counting belongs in metrics; security and business audit trails are a separate, protected stream.

## Standard fields

| Field | Notes |
|---|---|
| `timestamp` | UTC, ISO 8601 with milliseconds |
| `level` | debug, info, warn, error |
| `message` | Constant string |
| `service`, `version`, `environment` | Identify the emitter |
| `correlation_id` | Request/work id, propagated |
| `trace_id`, `span_id` | When tracing is on |
| `operation` | Use case or handler name |
| `duration_ms` | For completed operations |
| `outcome` | success, failure, and a stable error code |
| `actor_id` | Opaque internal id only, never a name or email |

## Levels

- **error**: the operation failed and needs attention; each error line has an error code and enough context to find it.
- **warn**: unexpected but handled (retry succeeded, degraded mode, rejected input from a client).
- **info**: significant lifecycle and business events (start, stop, request completed, message processed). Keep volume low; one line per request at most.
- **debug**: detail for diagnosis; off by default, switchable at runtime per component.

## What to log

- Start and stop of the service, configuration summary without secrets.
- One completion line per request or message: operation, outcome, duration, ids.
- Every error with its code, the failing dependency, and the retry decision.
- State transitions of long-running or asynchronous work, including retries, dead-letter routing, and reprocessing.
- Security-relevant decisions go to the audit stream (`security.md`).

## What never to log

Passwords, tokens, keys, full payment or account numbers, personal data, entire request or response bodies, raw exceptions from third parties that may contain secrets (log the code and a sanitised message).

## Volume, cost, and cardinality

- Log volume is a cost and a risk; budget it per service and review it.
- Sample high-volume info and debug lines; never sample errors.
- Keep fields bounded: do not put unbounded values (user-supplied strings, full URLs with ids) in indexed fields or metric labels.
- Retention is defined per stream (diagnostic short, audit longer) and written in the spec.

## Logs, metrics, traces

| Question | Tool |
|---|---|
| How many, how fast, how often? (SLIs, alerts) | Metrics |
| Where did the time go in this request? | Traces |
| What exactly happened in this case? | Logs |

Alerts are built on metrics (and synthetic probes), not on log text. See `alerting.md`. A log-based metric is acceptable when the signal exists only in logs, with a stable event name.

## Testing

- A test asserts that error paths emit the expected event with the error code.
- A test asserts that a known secret value does not appear in captured logs.
- The correlation id is present across a service boundary in integration tests.

## Review checklist

- [ ] Structured, constant messages, standard fields present?
- [ ] Correlation id created at the edge and propagated, including through messages?
- [ ] Errors logged once, at the boundary, with a code?
- [ ] No secrets or personal data possible in any field?
- [ ] Levels sensible; volume and retention defined?
