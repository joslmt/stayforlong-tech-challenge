# ADR 0002: Use weighted interval scheduling

- **Status:** Accepted
- **Date:** 2026-08-20

## Context

The revenue endpoint must handle large inputs and maximise occupied nights across conflicting intervals. Greedy arrival-order acceptance is fast but does not maximise the requested objective.

## Decision

Implement a domain `ScheduleOptimizer` service. Sort bookings by checkout date, find each compatible predecessor with binary search, and use dynamic programming to maximise occupied nights. Break ties by exact profit and then earliest input position. Reconstruct a validated `Schedule` and return it chronologically.

## Consequences

Runtime is `O(n log n)` and memory is `O(n)` for up to 10,000 candidates. The algorithm is less trivial than greedy selection, so focused unit tests explain the objective and tie-breakers.
