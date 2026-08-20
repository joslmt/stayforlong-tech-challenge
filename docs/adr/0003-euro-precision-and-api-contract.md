# ADR 0003: Preserve exact EUR calculations

- **Status:** Accepted
- **Date:** 2026-08-20

## Context

Published API versions disagree between integer and two-decimal results. Dividing a booking's profit by its nights often creates fractional cents.

## Decision

Follow the supplied decimal contract. Convert whole-euro selling rates and integer margins into exact profit cents, keep per-night values as rational numbers, aggregate before rounding, and round half-up to cents only at the response boundary.

The domain represents each rational value with `ProfitPerNight` and groups the three statistics in the `NightMetrics` composite value object. Application results carry those exact values unchanged. Only the HTTP adapter maps them to rounded JSON numbers.

## Consequences

Results are deterministic and avoid binary floating-point arithmetic in business decisions. JSON responses remain numbers, with no more than two decimal places. The contract divergence is documented separately.
