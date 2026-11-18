# Architecture Decision Records

Numbered, one decision each, written from the pet-project ADR template (scaffold with the pet-project skill `/new-adr`). Statuses: Proposed, Accepted, Rejected, Superseded, Deprecated. **Only the owner moves an ADR to Accepted.**

| ID | Title | Status |
|---|---|---|
| 0001 | Record architecture decisions | Proposed |
| 0002 | Docs-first workflow with an automated gate | Proposed |
| 0003 | Use Go for the expense ledger | Proposed |
| 0004 | gRPC with Protobuf, contract-first | Proposed |
| 0005 | PostgreSQL as the primary store | Proposed |
| 0006 | Kafka as the event backbone | Proposed |
| 0007 | Transactional outbox moved by Debezium CDC | Proposed |
| 0008 | Deploy on GCP Cloud Run with Terraform and workload identity | Proposed |

Decisions still to make (each becomes an ADR before the spec that depends on it): Kafka hosting option, Debezium hosting option, outbox payload format, Kafka client library, Postgres driver and query approach, observability stack.

The table above is maintained by hand; `python3 -I tools/docgate.py status` (run from the pet-project kit) is the source of truth for statuses.
