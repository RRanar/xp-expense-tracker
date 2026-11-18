---
id: 0012
title: Self-issued JWT access tokens with rotating refresh tokens, built to be extractable as an OAuth 2.0 service
status: Proposed
date: 2026-11-18
supersedes: none
---

# ADR 0012: Self-issued JWT access tokens with rotating refresh tokens, built to be extractable as an OAuth 2.0 service

## Context

The API needs to authenticate users and isolate their data. The owner wants the project to show how authentication is built, not delegate it to a hosted provider, and wants the design to allow the token logic to be extracted later into a stand-alone OAuth 2.0 token service without changing the resource API. Best-practice references to confirm when writing the milestone 4 spec: OAuth 2.0 Security Best Current Practice (RFC 9700), JWT profile for access tokens (RFC 9068), JWT best current practices (RFC 8725), and the OAuth 2.1 direction (no password grant, PKCE for public clients). These have not been re-read for this ADR, so this decision is unproven until the verification gate below passes.

## Decision

We will issue our own tokens: a short-lived JWT access token and an opaque, rotating refresh token, with a clean seam so the issuer can become a separate service.

- **Separate bounded context.** An `identity` context (own domain, use cases, persistence tables, and its own gRPC service in a separate proto package) issues and revokes tokens. The expense API is a resource server: it only validates access tokens through a small `TokenVerifier` port, and never reads identity tables.
- **Access token.** Signed JWT using an asymmetric algorithm (EdDSA or ES256, fixed in the spec), `kid` header, public keys published as a JWKS document, key rotation supported. Claims: `iss`, `sub` (user id), `aud` (the API), `exp`, `iat`, `jti`, and `scope`. Lifetime is short (default 10 to 15 minutes, owner tunes). The verifier pins the allowed algorithm, checks `iss`, `aud`, `exp` with small clock skew, and rejects `none` and algorithm confusion. The resource server holds only public keys.
- **Refresh token.** Opaque random value (not a JWT), at least 128 bits of entropy, stored only as a hash. Single use with **rotation**: each refresh returns a new pair. Reuse of a rotated token is treated as theft and revokes the whole token family, except that a short grace window lets a legitimate retry after a network failure succeed without being treated as theft (window length fixed in the spec). Idle and absolute lifetimes are spec defaults the owner tunes. Logout and password change revoke the family. The first clients are a command-line client and `grpcurl`, so the refresh token is kept in a user-only local file, which the spec's threat model covers.
- **Token endpoint semantics.** The identity gRPC service mirrors OAuth 2.0 (token request with a `refresh_token` grant, revocation, JWKS) so a later move to a real OAuth 2.0 authorization server does not change what resource servers verify. The first version offers a minimal local login (email and password, password hashed with argon2id) for synthetic accounts only, with no password reset and no email verification; these limits are stated in the README. A different flow can replace it in the milestone 4 spec if the owner prefers. The password grant is not copied from OAuth 2.0 as a protocol; it stays an internal login RPC.
- **Transport.** Tokens travel in gRPC metadata (`authorization: Bearer`) over TLS only. Refresh tokens are never sent to the resource API.
- **Authorisation.** Scopes are coarse (the list is fixed in the milestone 4 spec). Object-level ownership is enforced in the repository layer by `sub`.
- **Abuse controls.** Rate limit and lockout on login and refresh, uniform error messages, constant-time comparisons.

## Verification gate

Before the milestone 4 spec is approved, each practice above is checked against the current text of RFC 9700, RFC 9068 and RFC 8725 and the result is recorded in the spec. Until then this ADR is a proposal, not an audited design.

## Alternatives considered

- Hosted identity provider (Google Identity Platform or similar): least code and a strong security posture, but it hides the part the owner wants to show and adds a vendor dependency.
- Opaque access tokens with introspection: easy revocation, but every call needs a lookup, which is a poorer fit for decoupled resource servers.
- Long-lived JWT access tokens without refresh: simplest, but a stolen token stays useful for too long and cannot be revoked.
- Embedded identity in the API with shared tables: fastest to write, but breaks the extraction goal.

## Consequences

Easier: stateless verification in the API, a clear boundary, rotation and reuse detection limit the damage of theft, and the extraction path is mapped. Harder: the owner now owns a security-critical component (key management, rotation, password storage, revocation); it is a pet project with synthetic data and must be described that way, never as audited or production-hardened. Access tokens cannot be revoked before they expire, which is accepted because they are short-lived.

## Operational impact

Signing keys are generated by the identity service, stored in Secret Manager and rotated on a schedule. The previous public key stays in the JWKS for an overlap window at least as long as the access-token lifetime, so tokens already issued still verify. An emergency rotation removes the compromised key from the JWKS at once and accepts that tokens signed with it stop working. Alerts and runbooks: `AuthAbuseSuspected` (RFC 0001) covers login and refresh failure spikes and reuse-detection spikes; a further alert covers signing-key expiry; each has a runbook. Logs never contain tokens, token hashes or passwords. Drills: rotate the key with traffic running; replay a used refresh token and confirm the family is revoked.

Decoupling test: the expense API's tests pass using only a fake `TokenVerifier` and a fixed JWKS, and a dependency-rule check in CI fails if the API imports any identity package. The extraction itself (separate deployable, shared nothing but the JWKS and token contract) is not part of the first release.
