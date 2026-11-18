# Standard: security

Platform-agnostic. Security is designed in the spec, verified in review, and checked in CI. This is a baseline, not a certification.

## Threat modelling (per spec or RFC)

Keep it short and concrete. In the "Security" section of every spec/RFC:

1. **Assets**: what is worth protecting (accounts, personal or financial data, availability).
2. **Trust boundaries and entry points**: where untrusted input arrives (APIs, messages, files, operators).
3. **Threats** using STRIDE per boundary: Spoofing, Tampering, Repudiation, Information disclosure, Denial of service, Elevation of privilege.
4. **Mitigations** for each, mapped to a test or a control.
5. **Residual risk** accepted, with the reason.

## Authentication and authorisation

- Authenticate every request at the edge; never trust identity supplied by the caller in the body.
- **Authorise on every operation, deny by default.** Check object-level access (this user may touch this record), not only role.
- Short-lived credentials, signature and audience validation, clock-skew limits, revocation strategy.
- Service-to-service calls are authenticated too (workload identity, mutual authentication), with least privilege per caller.

## Input and output

- Validate all input at the boundary against an explicit schema: type, size, range, format. Reject, do not sanitise silently.
- Parameterised queries only; never build queries or commands by string concatenation.
- Limit payload sizes, collection sizes, pagination depth, and recursion.
- Output encoding for the target context; errors expose a stable code and a safe message, never internals or stack traces.

## Secrets and keys

- No secrets in source, configuration files in the repo, images, logs, or prompts. Use a secret manager and inject at runtime.
- Separate credentials per environment and per service; rotate; prefer short-lived and federated identities over long-lived keys.
- Secret scanning in CI and pre-commit.

## Least privilege

- Services, databases, pipelines and people get only the permissions they need, scoped narrowly and reviewed.
- Network access is allow-listed; databases are not publicly reachable.

## Data protection

- Classify data (public, internal, confidential, personal/financial). The spec states the class of each field.
- Encrypt in transit and at rest; minimise what is collected and retained; define retention and deletion.
- Personal data never appears in logs, metrics labels, traces or error messages (`logging.md`).
- Test and demo data are synthetic.

## Abuse and resilience

- Rate limiting and quotas per identity; request timeouts; bounded queues and retries with backoff and jitter.
- Idempotency for operations that can be retried.
- Graceful degradation under dependency failure; no unbounded resource use from a single caller.

## Supply chain

- Pin dependencies; review additions; scan for known vulnerabilities and licence issues on every build.
- Build from clean, reproducible environments; sign or verify artefacts where supported; generate an SBOM for released artefacts.
- Base images and tools are minimal and kept current.

## Audit and detection

- Security-relevant events (authentication, authorisation failures, permission changes, data export) are logged to an audit stream separate from diagnostics, with who, what, when, from where, and outcome.
- Alert on anomalies that suggest abuse (spikes of authorisation failures, impossible usage patterns). See `alerting.md`.

## Verification

- Security acceptance criteria are tests: unauthorised access is rejected, cross-tenant access is rejected, oversized and malformed input is rejected, secrets are absent from logs.
- Static analysis, dependency scan, secret scan, and fuzzing of parsers where relevant run in CI.
- Reference lists to consult during design: OWASP ASVS and the OWASP API Security Top 10.

## Review checklist

- [ ] Threat model present with STRIDE per boundary?
- [ ] Authorisation checked per object, deny by default?
- [ ] Input validated at the boundary with limits?
- [ ] Secrets externalised; none in logs?
- [ ] Data classified; retention defined?
- [ ] Rate limits, timeouts, idempotency where retries happen?
- [ ] Dependency, secret and static scans in CI?
- [ ] Security tests exist for each mitigation?
