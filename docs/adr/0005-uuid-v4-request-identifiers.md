# ADR 0005: Require UUID v4 booking request identifiers

- Status: accepted
- Date: 2026-08-20

## Context

The supplied contract describes `request_id` as a unique string and shows partner-prefixed values. The project needs one unambiguous identifier constraint that can be validated consistently across integrations.

## Decision

Every booking `request_id` must use the canonical hyphenated UUID v4 representation. Hexadecimal characters are accepted in either case, normalised to lowercase by the value object, and compared case-insensitively for payload uniqueness.

Validation checks both UUID semantics encoded in the text: the version nibble is `4`, and the RFC 4122 variant nibble is one of `8`, `9`, `a`, or `b`.

## Consequences

- Invalid, non-v4, and non-canonical identifiers return `400` with `request_id must be a valid UUID v4`.
- The API, frontend examples, README, tests, and OpenAPI schema use UUID v4 values.
- UUIDs that differ only by hexadecimal letter case are duplicates.
- This intentionally diverges from the free-form upstream examples and is recorded in the OAS inconsistency document.
