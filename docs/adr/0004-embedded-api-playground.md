# ADR 0004: Embed a focused API playground

- **Status:** Accepted
- **Date:** 2026-08-20

## Context

Reviewers need a quick way to exercise both endpoints, while curl examples and OpenAPI remain important. A separate frontend toolchain would add installation, supply-chain, and runtime overhead.

## Decision

Serve a dependency-free HTML/CSS/JavaScript playground from the Go binary. Model the interaction after Swagger UI: documented endpoint panels, editable examples, Execute, formatted responses, generated curl, and an FAQ tab. Use Stayforlong-pink styling and the title “Stayforlong Tech Challenge”.

## Consequences

One process and one Docker container deliver the full demonstration. The UI is intentionally not a general OpenAPI renderer; `api/openapi.yaml` remains the machine-readable contract.
