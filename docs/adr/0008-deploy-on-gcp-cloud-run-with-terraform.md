---
id: 0008
title: Deploy on GCP Cloud Run with Terraform and workload identity
status: Proposed
date: 2026-11-18
supersedes: none
---

# ADR 0008: Deploy on GCP Cloud Run with Terraform and workload identity

## Context

The project should show a deployed service on a major cloud with infrastructure as code and secure delivery, at low cost. The owner has Kubernetes, Helm and ArgoCD experience from work; Cloud Run is the cheaper, simpler first target.

## Decision

We will deploy the services as containers on Google Cloud Run, with infrastructure defined in Terraform, images in Artifact Registry, secrets in Secret Manager, and the delivery pipeline authenticating through workload identity federation (no long-lived keys). A long-lived consumer needs always-on instances (minimum one, CPU always allocated) or another runtime; this and gRPC/HTTP2 settings are to be verified in the current Cloud Run documentation when the deployment spec is written. A Kubernetes variant (Helm, GitOps) is optional, after the first deployment works.

## Alternatives considered

- GKE Autopilot from the start: reuses existing Kubernetes skills and suits long-lived consumers, but costs more and adds setup before anything is deployed.
- A virtual machine: cheapest to reason about, but manual operations and weak demonstration of cloud-native practice.
- Another cloud: no advantage for the stated goal, and the owner's target roles name GCP.

## Consequences

Easier: scale to zero for the API, small ops surface, quick deploys. Harder: consumers and CDC components do not fit scale-to-zero, so part of the system runs continuously and costs money; some features need verification against current platform limits.

## Operational impact

Per-service identities with least privilege; a billing budget with alerts and a teardown command before any resource is created; logs and traces exported to the platform's observability services following `pet-project/standards/logging`; availability and latency SLIs from the platform and application metrics feed the alerts in `pet-project/standards/alerting`. Costs of managed database and broker options must be checked before creation.
