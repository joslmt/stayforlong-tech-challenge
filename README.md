# Stayforlong Tech Challenge

A small booking-profit API built to be easy to run, inspect, and trust. It exposes the two requested endpoints, includes an interactive API playground, and documents the business decisions that the challenge leaves open.

## Start here

You only need Docker Desktop (or Docker Engine with Compose) and Make:

```bash
make start
```

Then open **[http://localhost:8080](http://localhost:8080)**. The **Stayforlong Tech Challenge** frontend works like a focused Swagger UI: choose an endpoint, edit the prefilled JSON, execute it, inspect the response, or copy the generated curl command. The **FAQs** tab explains the business rules in plain language.

If you prefer the terminal, these examples are ready to copy.

### `/data`

```bash
curl --request POST 'http://localhost:8080/data' \
  --header 'Content-Type: application/json' \
  --data-raw '[
    {
      "request_id": "550e8400-e29b-41d4-a716-446655440001",
      "check_in": "2026-09-01",
      "nights": 2,
      "selling_rate": 240,
      "margin": 20
    },
    {
      "request_id": "550e8400-e29b-41d4-a716-446655440002",
      "check_in": "2026-09-05",
      "nights": 3,
      "selling_rate": 450,
      "margin": 18
    }
  ]'
```

### `/revenue`

```bash
curl --request POST 'http://localhost:8080/revenue' \
  --header 'Content-Type: application/json' \
  --data-raw '[
    {
      "request_id": "550e8400-e29b-41d4-a716-446655440001",
      "check_in": "2026-09-01",
      "nights": 4,
      "selling_rate": 480,
      "margin": 20
    },
    {
      "request_id": "550e8400-e29b-41d4-a716-446655440002",
      "check_in": "2026-09-05",
      "nights": 3,
      "selling_rate": 390,
      "margin": 18
    },
    {
      "request_id": "550e8400-e29b-41d4-a716-446655440003",
      "check_in": "2026-09-03",
      "nights": 6,
      "selling_rate": 780,
      "margin": 15
    }
  ]'
```

Stop the project with `make stop`. The Makefile delegates to Docker Compose, so the same application stack runs on macOS, Linux, and Windows environments that provide Make. If Make is unavailable, the equivalent fallback is `docker compose up --build`.

If port `8080` is occupied, choose another loopback port without changing the application:

```bash
make start PORT=8081
```

Then use `http://localhost:8081`. Docker Compose binds the selected port to `127.0.0.1`, so the challenge is not exposed to the local network by default.

## What the API does

`POST /data` calculates average, minimum, and maximum profit per night across **all** supplied booking requests.

`POST /revenue` finds a non-conflicting schedule that maximises **occupied nights**. Ties prefer higher total profit and then the earliest submitted booking. Returned request IDs are ordered chronologically, an explicit choice documented because one supplied example appears to retain input order.

The apartment operates with check-in at 15:00, checkout at 11:00, and cleaning from 11:00 to 12:00. A stay can start on the previous stay's checkout date.

All amounts are euros. The request uses whole EUR values for `selling_rate`. Calculations remain exact internally in cents/rational values. **Every exposed EUR metric is rounded half-up to two decimal places only at the response boundary**; intermediate nightly values are never rounded and summed.

Every `request_id` must be a canonical hyphenated UUID v4. UUID hexadecimal input is case-insensitive, normalised to lowercase, and must be unique within the payload. Values using another UUID version—or partner-style identifiers such as `bookata_XY123`—return a human-friendly `400` response.

See [Business decisions](docs/business-decisions.md) for every validation rule and [OAS inconsistencies](docs/oas-inconsistencies.md) for differences between the published API definitions.

## Architecture

The repository uses a deliberately small hexagonal architecture around one **Booking Optimisation** bounded context:

```text
HTTP / JSON adapter ──> application use cases ──> booking domain
       │                                               │
       └── embedded API playground              pure business rules
```

The inbound HTTP adapter lives at `internal/booking/adapters/in/http`. Dependency wiring and booking-specific routes live in `internal/bootstrap`; reusable response-safety and observability middleware live in `internal/platform/httpserver`; process lifecycle remains in `cmd/api`.

- `BookingRequest` is an immutable domain entity composed of validated value objects.
- `BookingBatch` is the request aggregate boundary: it enforces cardinality and unique IDs regardless of which adapter invokes the application.
- `ProfitPerNight` is an exact rational value object, and `NightMetrics` groups average, minimum, and maximum without introducing `float64` into the domain.
- `Schedule` is the aggregate that protects the no-overlap invariant.
- `ScheduleOptimizer` is the domain service that applies weighted interval scheduling and business tie-breakers.
- `MetricsService` and `ScheduleService` are small application services that orchestrate domain behavior.
- HTTP DTOs and human-friendly error mapping stay in the inbound adapter.
- There is no repository port or database: both results depend only on the current request. Adding persistence would create accidental complexity rather than demonstrate architecture.

The scheduler is weighted interval scheduling: sort by checkout, binary-search compatible predecessors, run dynamic programming, then reconstruct the chosen bookings. Complexity is `O(n log n)` time and `O(n)` memory for up to 10,000 requests.

Meaningful architecture choices are recorded under [`docs/adr`](docs/adr).

## Error handling

Domain and application failures use stable numeric codes internally, such as `1102` for duplicate IDs. Codes and technical details are written to structured logs. The public contract intentionally remains:

```json
{
  "message": "Human-friendly explanation"
}
```

Validation failures return `400`. Unexpected errors and recovered panics return a safe `500` without exposing internals. Every request receives an `X-Request-ID`; context cancellation, body limits, server timeouts, and graceful shutdown are handled explicitly.

`GET /healthz` returns an empty `204` for container health probes. Successful probes are deliberately omitted from access logs. The healthcheck URL can be overridden with `HEALTHCHECK_URL` when running the binary outside Docker.

The complete internal mapping is documented in [`docs/error-codes.md`](docs/error-codes.md).

## Develop and test

Local development requires Go 1.25 or later; the repository pins the current Go 1.27.0 toolchain used by Docker and CI:

```bash
go run ./cmd/api
go test -race ./...
```

Useful commands:

```bash
make tests         # unit and integration tests, race detector, production coverage
make e2e           # Testcontainers black-box tests against the production image
make fmt-check     # verify formatting
make vet           # Go static analysis
make vuln          # Go vulnerability scan (requires govulncheck)
```

Tests use Object Mothers to keep fixtures semantic and Given-When-Then structure where it makes behavior easier to read. The suite covers value-object constraints, exact money arithmetic, schedule optimisation and tie-breakers, 10,000 inputs, HTTP contracts, safe errors, and frontend delivery. E2E tests use [Testcontainers for Go](https://github.com/testcontainers/testcontainers-go) to build the real multi-stage Dockerfile, start an isolated production container, discover its mapped port, exercise the complete API and frontend, and clean everything up from `go test`—the exact same command used in CI.

`make tests` reports aggregate executable-statement coverage for production packages rather than mixing in test fixture helpers and development tooling. CI separately requires at least 90% coverage for the core domain/application scope and the HTTP integration scope. Coverage is a regression guard, not a substitute for meaningful assertions.

## Delivery

- Multi-stage, non-root, distroless Docker image with a read-only filesystem and dropped Linux capabilities.
- Separate GitHub Actions jobs for formatting, linting, Go vulnerabilities, unit/integration tests, OpenAPI validation, image security/SBOM, Docker build, and E2E tests. Third-party actions are pinned to immutable commits, and the security job scans the exact image artifact produced by the build job.
- Dependency updates managed through Dependabot.

The complete API contract is [`api/openapi.yaml`](api/openapi.yaml).
