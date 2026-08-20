# ADR 0001: Keep the booking API stateless

- **Status:** Accepted
- **Date:** 2026-08-20

## Context

Both endpoints receive the complete data set needed to calculate their response. The challenge asks for architectural quality but does not ask to store bookings.

## Decision

Use one Booking Optimisation bounded context with domain, application, and inbound HTTP layers. Do not add a repository interface or database. Interfaces appear at the inbound application boundary so transport adapters depend on use-case contracts rather than concrete services.

Place transport mapping under `internal/booking/adapters/in/http` to make its inbound direction explicit. Put process composition and booking-specific routes under `internal/bootstrap`; keep reusable HTTP middleware under `internal/platform/httpserver`; keep server lifecycle under `cmd/api`.

Model a request as a `BookingBatch` aggregate so cardinality and uniqueness rules survive every adapter. Keep weighted scheduling in the domain `ScheduleOptimizer`; application services orchestrate that behavior and shape use-case results.

## Consequences

Business behavior remains testable without HTTP, while the project stays small enough to understand. Horizontal instances remain independent. Persistence can be introduced behind a new outbound port if a future use case actually requires it.
